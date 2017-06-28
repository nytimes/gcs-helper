package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/fsouza/fake-gcs-server/fakestorage"
)

func TestServerMultiPrefixes(t *testing.T) {
	addr, cleanup := startServer(t, Config{
		BucketName:    "my-bucket",
		MapPrefix:     "/map/",
		ProxyPrefix:   "/proxy/",
		MapExtensions: []string{".mp3", ".txt"},
	})
	defer cleanup()
	var tests = []serverTest{
		{
			testCase:       "healthcheck",
			method:         "GET",
			addr:           addr,
			expectedStatus: http.StatusOK,
		},
		{
			testCase:       "not found",
			method:         "GET",
			addr:           addr + "/what",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "not found\n",
		},
		{
			testCase:       "proxy: download file",
			method:         "GET",
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
			method:   "GET",
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
			method:         "GET",
			addr:           addr + "/map/musics/music/mu",
			expectedStatus: http.StatusOK,
			expectedHeader: http.Header{"Content-Type": []string{"application/json"}},
			expectedBody: map[string]interface{}{
				"sequences": []interface{}{
					map[string]interface{}{
						"clips": []interface{}{
							map[string]interface{}{
								"type": "source",
								"path": "/musics/music/music1.txt",
							},
						},
					},
					map[string]interface{}{
						"clips": []interface{}{
							map[string]interface{}{
								"type": "source",
								"path": "/musics/music/music2.txt",
							},
						},
					},
					map[string]interface{}{
						"clips": []interface{}{
							map[string]interface{}{
								"type": "source",
								"path": "/musics/music/music3.txt",
							},
						},
					},
					map[string]interface{}{
						"clips": []interface{}{
							map[string]interface{}{
								"type": "source",
								"path": "/musics/music/music4.mp3",
							},
						},
					},
				},
			},
		},
		{
			testCase:       "map: empty list",
			method:         "GET",
			addr:           addr + "/map/musics/musyc",
			expectedStatus: http.StatusOK,
			expectedHeader: http.Header{"Content-Type": []string{"application/json"}},
			expectedBody:   map[string]interface{}{"sequences": []interface{}{}},
		},
		{
			testCase:       "map: method not allowed",
			method:         "POST",
			addr:           addr + "/map/musics",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "method not allowed\n",
		},
		{
			testCase:       "map: invalid url",
			method:         "GET",
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

func startServer(t *testing.T, cfg Config) (string, func()) {
	server := fakestorage.NewServer(getObjects())
	handler := getHandler(cfg, server.Client())
	httpServer := httptest.NewServer(handler)
	return httpServer.URL, func() {
		httpServer.Close()
		server.Stop()
	}
}

func getObjects() []fakestorage.Object {
	return []fakestorage.Object{
		{
			BucketName: "my-bucket",
			Name:       "musics/music/music1.txt",
			Content:    []byte("some nice music"),
		},
		{
			BucketName: "my-bucket",
			Name:       "musics/music/music2.txt",
			Content:    []byte("some nicer music"),
		},
		{
			BucketName: "my-bucket",
			Name:       "musics/music/music3.txt",
			Content:    []byte("some even nicer music"),
		},
		{
			BucketName: "my-bucket",
			Name:       "musics/music/music4.mp3",
		},
		{
			BucketName: "my-bucket",
			Name:       "musics/music/music5.wav",
		},
		{
			BucketName: "my-bucket",
			Name:       "musics/music/music/1.txt",
		},
		{
			BucketName: "my-bucket",
			Name:       "musics/music/music/2.txt",
		},
		{
			BucketName: "my-bucket",
			Name:       "musics/music/music/3.txt",
		},
		{
			BucketName: "my-bucket",
			Name:       "musics/music/music/4.mp3",
		},
		{
			BucketName: "my-bucket",
			Name:       "musics/musics/music1.txt",
		},
		{
			BucketName: "your-bucket",
			Name:       "musics/music/music3.txt",
			Content:    []byte("wait what"),
		},
	}
}
