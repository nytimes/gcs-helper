package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestHTTPClient(t *testing.T) {
	hc := httpClient(ClientConfig{
		Timeout:         time.Minute,
		IdleConnTimeout: 2 * time.Minute,
		MaxIdleConns:    10,
	})
	expectedClient := http.Client{
		Timeout: time.Minute,
		Transport: &http.Transport{
			MaxIdleConns:    10,
			IdleConnTimeout: 2 * time.Minute,
		},
	}
	ign := cmpopts.IgnoreUnexported(http.Transport{})
	if !cmp.Equal(*hc, expectedClient, ign) {
		t.Errorf("wrong client returned\n%s", cmp.Diff(*hc, expectedClient, ign))
	}
}
