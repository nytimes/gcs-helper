package handlers

import (
	"net/http/httptest"

	"github.com/NYTimes/gcs-helper/v3/internal/testhelper"
	"github.com/fsouza/fake-gcs-server/fakestorage"
)

func testMapServer(cfg Config) (string, func()) {
	server := fakestorage.NewServer(testhelper.FakeObjects)
	httpServer := httptest.NewServer(Map(cfg, server.Client()))
	return httpServer.URL, func() {
		httpServer.Close()
		server.Stop()
	}
}

func testProxyServer(cfg Config) (string, func()) {
	server := fakestorage.NewServer(testhelper.FakeObjects)
	httpServer := httptest.NewServer(Proxy(cfg, server.Client()))
	return httpServer.URL, func() {
		httpServer.Close()
		server.Stop()
	}
}
