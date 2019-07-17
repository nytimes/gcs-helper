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
	tests := []testhelper.ServerTest{
		{
			TestCase:       "healthcheck through the proxy",
			Method:         http.MethodGet,
			Addr:           addr,
			ReqHeader:      nil,
			ExpectedStatus: http.StatusOK,
			ExpectedHeader: nil,
			ExpectedBody:   "",
		},
		{
			TestCase:       "download file",
			Method:         http.MethodGet,
			Addr:           addr + "/musics/music/music1.txt",
			ReqHeader:      nil,
			ExpectedStatus: http.StatusOK,
			ExpectedHeader: http.Header{
				"Accept-Ranges":  []string{"bytes"},
				"Content-Length": []string{"15"},
			},
			ExpectedBody: "some nice music",
		},
		{
			TestCase: "download file - range",
			Method:   http.MethodGet,
			Addr:     addr + "/musics/music/music2.txt",
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
			TestCase:       "file attrs",
			Method:         http.MethodHead,
			Addr:           addr + "/musics/music/music2.txt",
			ReqHeader:      nil,
			ExpectedStatus: http.StatusOK,
			ExpectedHeader: http.Header{
				"Accept-Ranges":  []string{"bytes"},
				"Content-Length": []string{"16"},
			},
			ExpectedBody: "",
		},
		{
			TestCase:       "download file - object not found",
			Method:         http.MethodGet,
			Addr:           addr + "/musics/music/some-music.txt",
			ReqHeader:      nil,
			ExpectedStatus: http.StatusNotFound,
			ExpectedHeader: nil,
			ExpectedBody:   "not found\n",
		},
		{
			TestCase:       "file attrs - object not found",
			Method:         http.MethodHead,
			Addr:           addr + "/musics/music/some-music.txt",
			ReqHeader:      nil,
			ExpectedStatus: http.StatusNotFound,
			ExpectedHeader: nil,
			ExpectedBody:   "",
		},
		{
			TestCase:       "method not allowed - POST",
			Method:         http.MethodPost,
			Addr:           addr + "/whatever",
			ReqHeader:      nil,
			ExpectedStatus: http.StatusMethodNotAllowed,
			ExpectedHeader: nil,
			ExpectedBody:   "method not allowed\n",
		},
		{
			TestCase:       "method not allowed - PUT",
			Method:         http.MethodPut,
			Addr:           addr + "/whatever",
			ReqHeader:      nil,
			ExpectedStatus: http.StatusMethodNotAllowed,
			ExpectedHeader: nil,
			ExpectedBody:   "method not allowed\n",
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
	tests := []testhelper.ServerTest{
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
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("wrong status code\nwant %d\ngot  %d", http.StatusNotFound, resp.StatusCode)
	}
}
