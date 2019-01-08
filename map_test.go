package main

import (
	"net/http"
	"testing"
	"time"
)

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
		t.Errorf("wrong status code\nwant %d\ngot  %d", http.StatusNotFound, resp.StatusCode)
	}
}
