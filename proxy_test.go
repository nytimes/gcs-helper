package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fsouza/fake-gcs-server/fakestorage"
)

func TestServer(t *testing.T) {
	addr, cleanup := startServer(t, Config{BucketName: "my-bucket"})
	defer cleanup()
	var tests = []struct {
		testCase       string
		method         string
		path           string
		reqHeader      http.Header
		expectedStatus int
		expectedHeader http.Header
		expectedBody   string
	}{
		{
			"healthcheck",
			"GET",
			"/",
			nil,
			http.StatusOK,
			nil,
			"",
		},
		{
			"download file",
			"GET",
			"/musics/music/music1.txt",
			nil,
			http.StatusOK,
			http.Header{
				"Accept-Ranges":  []string{"bytes"},
				"Content-Length": []string{"15"},
			},
			"some nice music",
		},
		{
			"download file - range",
			"GET",
			"/musics/music/music2.txt",
			http.Header{
				"Range": []string{"bytes=2-10"},
			},
			http.StatusPartialContent,
			http.Header{
				"Accept-Ranges":  []string{"bytes"},
				"Content-Length": []string{"8"},
				"Content-Range":  []string{"bytes 2-10/16"},
			},
			"me nicer",
		},
		{
			"file attrs",
			"HEAD",
			"/musics/music/music2.txt",
			nil,
			http.StatusOK,
			http.Header{
				"Accept-Ranges":  []string{"bytes"},
				"Content-Length": []string{"16"},
			},
			"",
		},
		{
			"download file - object not found",
			"GET",
			"/musics/music/some-music.txt",
			nil,
			http.StatusNotFound,
			nil,
			"storage: object doesn't exist\n",
		},
		{
			"file attrs - object not found",
			"HEAD",
			"/musics/music/some-music.txt",
			nil,
			http.StatusNotFound,
			nil,
			"",
		},
		{
			"method not allowed - POST",
			"POST",
			"/whatever",
			nil,
			http.StatusMethodNotAllowed,
			nil,
			"method not allowed\n",
		},
		{
			"method not allowed - PUT",
			"PUT",
			"/whatever",
			nil,
			http.StatusMethodNotAllowed,
			nil,
			"method not allowed\n",
		},
	}
	for _, test := range tests {
		t.Run(test.testCase, func(t *testing.T) {
			req, _ := http.NewRequest(test.method, addr+test.path, nil)
			for name := range test.reqHeader {
				req.Header.Set(name, test.reqHeader.Get(name))
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
			if resp.StatusCode != test.expectedStatus {
				t.Errorf("wrong status code\nwant %d\ngot  %d", test.expectedStatus, resp.StatusCode)
			}
			if string(data) != test.expectedBody {
				t.Errorf("wrong body\nwant %q\ngot  %q", test.expectedBody, string(data))
			}
			for name := range test.expectedHeader {
				expected := test.expectedHeader.Get(name)
				got := resp.Header.Get(name)
				if expected != got {
					t.Errorf("header %q: wrong value\nwant %q\ngot  %q", name, expected, got)
				}
			}
		})
	}
}

func TestServerProxyHandlerBucketNotFound(t *testing.T) {
	addr, cleanup := startServer(t, Config{BucketName: "some-bucket"})
	defer cleanup()
	req, _ := http.NewRequest("HEAD", addr+"/whatever", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("wrong status code\nwant %d\ngot  %d", http.StatusNotFound, resp.StatusCode)
	}
}

func startServer(t *testing.T, cfg Config) (string, func()) {
	server := fakestorage.NewServer(getObjects())
	handler := getProxyHandler(cfg, server.Client())
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
	}
}
