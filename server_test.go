package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/NYTimes/gcs-helper/v3/internal/testhelper"
	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/google/go-cmp/cmp"
)

func TestServerMultiPrefixes(t *testing.T) {
	addr, cleanup := startServer(Config{
		BucketName: "my-bucket",
		Map: MapConfig{
			Endpoint:    "/map/",
			RegexFilter: `((240|360|424|480|720|1080)p\.mp4)|\.(vtt)$`,
		},
		Proxy: ProxyConfig{
			Endpoint: "/proxy/",
			Timeout:  time.Second,
		},
	})
	defer cleanup()
	var tests = []serverTest{
		{
			testCase:       "healthcheck",
			method:         http.MethodGet,
			addr:           addr,
			expectedStatus: http.StatusOK,
		},
		{
			testCase:       "not found",
			method:         http.MethodGet,
			addr:           addr + "/what",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "not found\n",
		},
		{
			testCase:       "proxy: download file",
			method:         http.MethodGet,
			addr:           addr + "/proxy/musics/music/music1.txt",
			expectedStatus: http.StatusOK,
			expectedHeader: http.Header{
				"Accept-Ranges":  []string{"bytes"},
				"Content-Length": []string{"15"},
			},
			expectedBody: "some nice music",
		},
		{
			testCase: "proxy: download file - range",
			method:   http.MethodGet,
			addr:     addr + "/proxy/musics/music/music2.txt",
			reqHeader: http.Header{
				"Range": []string{"bytes=2-10"},
			},
			expectedStatus: http.StatusPartialContent,
			expectedHeader: http.Header{
				"Accept-Ranges":  []string{"bytes"},
				"Content-Length": []string{"8"},
				"Content-Range":  []string{"bytes 2-10/16"},
			},
			expectedBody: "me nicer",
		},
		{
			testCase:       "map: list of files",
			method:         http.MethodGet,
			addr:           addr + "/map/videos/video/",
			expectedStatus: http.StatusOK,
			expectedHeader: http.Header{"Content-Type": []string{"application/json"}},
			expectedBody: map[string]interface{}{
				"sequences": []interface{}{
					map[string]interface{}{
						"clips": []interface{}{
							map[string]interface{}{
								"type": "source",
								"path": "/my-bucket/videos/video/28043_1_video_1080p.mp4",
							},
						},
					},
					map[string]interface{}{
						"clips": []interface{}{
							map[string]interface{}{
								"type": "source",
								"path": "/my-bucket/videos/video/77071_1_caption_wg_240p_001f8ea7-749b-4d43-7bd5-b357e4e24f32.vtt",
							},
						},
					},
					map[string]interface{}{
						"clips": []interface{}{
							map[string]interface{}{
								"type": "source",
								"path": "/my-bucket/videos/video/video1_480p.mp4",
							},
						},
					},
					map[string]interface{}{
						"clips": []interface{}{
							map[string]interface{}{
								"type": "source",
								"path": "/my-bucket/videos/video/video1_720p.mp4",
							},
						},
					},
				},
			},
		},
		{
			testCase:       "map: empty list",
			method:         http.MethodGet,
			addr:           addr + "/map/musics/musyc",
			expectedStatus: http.StatusOK,
			expectedHeader: http.Header{"Content-Type": []string{"application/json"}},
			expectedBody:   map[string]interface{}{"sequences": []interface{}{}},
		},
		{
			testCase:       "map: method not allowed",
			method:         http.MethodPost,
			addr:           addr + "/map/musics",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "method not allowed\n",
		},
		{
			testCase:       "map: invalid url",
			method:         http.MethodGet,
			addr:           addr + "/map/",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "prefix cannot be empty\n",
		},
	}

	for _, test := range tests {
		t.Run(test.testCase, test.run)
	}
}

type serverTest struct {
	testCase       string
	method         string
	addr           string
	reqHeader      http.Header
	expectedStatus int
	expectedHeader http.Header
	expectedBody   interface{}
}

func (st *serverTest) run(t *testing.T) {
	req, _ := http.NewRequest(st.method, st.addr, nil)
	for name := range st.reqHeader {
		req.Header.Set(name, st.reqHeader.Get(name))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != st.expectedStatus {
		t.Errorf("wrong status code\nwant %d\ngot  %d", st.expectedStatus, resp.StatusCode)
	}
	if expectedBody, ok := st.expectedBody.(string); ok {
		if string(data) != expectedBody {
			t.Errorf("wrong body\nwant %q\ngot  %q", expectedBody, string(data))
		}
	} else if st.expectedBody != nil {
		var body interface{}
		err = json.Unmarshal(data, &body)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(body, st.expectedBody) {
			t.Log(cmp.Diff(body, st.expectedBody))
			t.Errorf("wrong body returned\nwant %#v\ngot  %#v", st.expectedBody, body)
		}
	}
	for name := range st.expectedHeader {
		expected := st.expectedHeader.Get(name)
		got := resp.Header.Get(name)
		if expected != got {
			t.Errorf("header %q: wrong value\nwant %q\ngot  %q", name, expected, got)
		}
	}
}

func startServer(cfg Config) (string, func()) {
	server := fakestorage.NewServer(testhelper.FakeObjects)
	handler := getHandler(cfg, server.Client())
	httpServer := httptest.NewServer(handler)
	return httpServer.URL, func() {
		httpServer.Close()
		server.Stop()
	}
}
