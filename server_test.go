package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/NYTimes/gcs-helper/v3/handlers"
	"github.com/NYTimes/gcs-helper/v3/internal/testhelper"
	"github.com/fsouza/fake-gcs-server/fakestorage"
)

func TestServerMultiPrefixes(t *testing.T) {
	addr, cleanup := startServer(t, handlers.Config{
		BucketName: "my-bucket",
		Map: handlers.MapConfig{
			Endpoint:    "/map/",
			RegexFilter: `((240|360|424|480|720|1080)p\.mp4)|\.(vtt)$`,
		},
		Proxy: handlers.ProxyConfig{
			Endpoint: "/proxy/",
			Timeout:  time.Second,
		},
	})
	defer cleanup()
	var tests = []testhelper.ServerTest{
		{
			TestCase:       "healthcheck",
			Method:         http.MethodGet,
			Addr:           addr,
			ExpectedStatus: http.StatusOK,
		},
		{
			TestCase:       "not found",
			Method:         http.MethodGet,
			Addr:           addr + "/what",
			ExpectedStatus: http.StatusNotFound,
			ExpectedBody:   "not found\n",
		},
		{
			TestCase:       "proxy: download file",
			Method:         http.MethodGet,
			Addr:           addr + "/proxy/musics/music/music1.txt",
			ExpectedStatus: http.StatusOK,
			ExpectedHeader: http.Header{
				"Accept-Ranges":  []string{"bytes"},
				"Content-Length": []string{"15"},
			},
			ExpectedBody: "some nice music",
		},
		{
			TestCase: "proxy: download file - range",
			Method:   http.MethodGet,
			Addr:     addr + "/proxy/musics/music/music2.txt",
			ReqHeader: http.Header{
				"Range": []string{"bytes=2-10"},
			},
			ExpectedStatus: http.StatusPartialContent,
			ExpectedHeader: http.Header{
				"Accept-Ranges":  []string{"bytes"},
				"Content-Length": []string{"8"},
				"Content-Range":  []string{"bytes 2-10/16"},
			},
			ExpectedBody: "me nicer",
		},
		{
			TestCase:       "map: list of files",
			Method:         http.MethodGet,
			Addr:           addr + "/map/videos/video/",
			ExpectedStatus: http.StatusOK,
			ExpectedHeader: http.Header{"Content-Type": []string{"application/json"}},
			ExpectedBody: map[string]interface{}{
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
			TestCase:       "map: empty list",
			Method:         http.MethodGet,
			Addr:           addr + "/map/musics/musyc",
			ExpectedStatus: http.StatusOK,
			ExpectedHeader: http.Header{"Content-Type": []string{"application/json"}},
			ExpectedBody:   map[string]interface{}{"sequences": []interface{}{}},
		},
		{
			TestCase:       "map: method not allowed",
			Method:         http.MethodPost,
			Addr:           addr + "/map/musics",
			ExpectedStatus: http.StatusMethodNotAllowed,
			ExpectedBody:   "method not allowed\n",
		},
		{
			TestCase:       "map: invalid url",
			Method:         http.MethodGet,
			Addr:           addr + "/map/",
			ExpectedStatus: http.StatusBadRequest,
			ExpectedBody:   "prefix cannot be empty\n",
		},
	}

	for _, test := range tests {
		t.Run(test.TestCase, test.Run)
	}
}

func startServer(t *testing.T, cfg handlers.Config) (string, func()) {
	server, err := fakestorage.NewServerWithOptions(fakestorage.Options{
		InitialObjects: testhelper.FakeObjects,
		NoListener:     true,
	})
	if err != nil {
		t.Fatal(err)
	}
	handler := getHandler(cfg, server.Client(), server.HTTPClient())
	httpServer := httptest.NewServer(handler)
	return httpServer.URL, func() {
		httpServer.Close()
		server.Stop()
	}
}
