package main

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/NYTimes/gcs-helper/v3/vodmodule"
	"github.com/google/go-cmp/cmp"
)

func TestServerMapListOfFiles(t *testing.T) {
	addr, cleanup := startServer(Config{
		BucketName: "my-bucket",
		Map: MapConfig{
			Endpoint:    "/map/",
			RegexFilter: `\.mp4$`,
		},
		Proxy: ProxyConfig{
			Endpoint: "/proxy/",
			Timeout:  time.Second,
		},
	})
	defer cleanup()
	req, _ := http.NewRequest(http.MethodGet, addr+"/map/videos/video/", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("wrong status code\nwant %d\ngot  %d", http.StatusOK, resp.StatusCode)
	}
	expectedBody := vodmodule.Mapping{
		Sequences: []vodmodule.Sequence{
			{
				Clips: []vodmodule.Clip{
					{
						Type: "source",
						Path: "/my-bucket/videos/video/28043_1_video_1080p.mp4",
					},
				},
			},
			{
				Clips: []vodmodule.Clip{
					{
						Type: "source",
						Path: "/my-bucket/videos/video/video1_480p.mp4",
					},
				},
			},
			{
				Clips: []vodmodule.Clip{
					{
						Type: "source",
						Path: "/my-bucket/videos/video/video1_720p.mp4",
					},
				},
			},
		},
	}
	var body vodmodule.Mapping
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(body, expectedBody) {
		t.Errorf("wrong body returned\nwant %#v\ngot  %#v\ndiff: %s", expectedBody, body, cmp.Diff(body, expectedBody))
	}
}

func TestServerMapBucketNotFound(t *testing.T) {
	addr, cleanup := startServer(Config{
		BucketName: "some-bucket",
		Map: MapConfig{
			Endpoint: "/map",
		},
		Proxy: ProxyConfig{
			Endpoint: "/proxy",
			Timeout:  time.Second,
		},
	})
	defer cleanup()
	req, _ := http.NewRequest(http.MethodGet, addr+"/map/whatever", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("wrong status code\nwant %d\ngot  %d", http.StatusInternalServerError, resp.StatusCode)
	}
}
