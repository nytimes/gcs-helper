package handlers

import (
	"net/http"
	"testing"
	"time"

	"github.com/NYTimes/gcs-helper/v3/internal/testhelper"
)

func TestProxyHandler(t *testing.T) {
	addr, cleanup := testProxyServer(t, Config{
		BucketName: "my-bucket",
		Proxy: ProxyConfig{
			LogHeaders: []string{"Accept", "User-Agent", "Range"},
			Timeout:    time.Second,
		},
	})
	defer cleanup()
	var tests = []testhelper.ServerTest{
		{
			"healthcheck through the proxy",
			http.MethodGet,
			addr,
			nil,
			http.StatusOK,
			nil,
			"",
		},
		{
			"download file",
			http.MethodGet,
			addr + "/musics/music/music1.txt",
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
			http.MethodGet,
			addr + "/musics/music/music2.txt",
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
			http.MethodHead,
			addr + "/musics/music/music2.txt",
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
			http.MethodGet,
			addr + "/musics/music/some-music.txt",
			nil,
			http.StatusNotFound,
			nil,
			"not found\n",
		},
		{
			"file attrs - object not found",
			http.MethodHead,
			addr + "/musics/music/some-music.txt",
			nil,
			http.StatusNotFound,
			nil,
			"",
		},
		{
			"method not allowed - POST",
			http.MethodPost,
			addr + "/whatever",
			nil,
			http.StatusMethodNotAllowed,
			nil,
			"method not allowed\n",
		},
		{
			"method not allowed - PUT",
			http.MethodPut,
			addr + "/whatever",
			nil,
			http.StatusMethodNotAllowed,
			nil,
			"method not allowed\n",
		},
	}
	for _, test := range tests {
		t.Run(test.TestCase, test.Run)
	}
}

func TestServerProxyHandlerBucketInThePath(t *testing.T) {
	addr, cleanup := testProxyServer(t, Config{
		BucketName: "my-bucket",
		Proxy: ProxyConfig{
			Endpoint:     "/proxy/",
			BucketOnPath: true,
			Timeout:      time.Second,
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
			TestCase:       "proxy: download file",
			Method:         http.MethodGet,
			Addr:           addr + "/your-bucket/musics/music/music3.txt",
			ExpectedStatus: http.StatusOK,
			ExpectedHeader: http.Header{
				"Content-Length": []string{"9"},
			},
			ExpectedBody: "wait what",
		},
	}

	for _, test := range tests {
		t.Run(test.TestCase, test.Run)
	}
}

func TestServerProxyHandlerBucketNotFound(t *testing.T) {
	addr, cleanup := testProxyServer(t, Config{BucketName: "some-bucket", Proxy: ProxyConfig{Timeout: time.Second}})
	defer cleanup()
	req, _ := http.NewRequest(http.MethodHead, addr+"/whatever", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("wrong status code\nwant %d\ngot  %d", http.StatusNotFound, resp.StatusCode)
	}
}
