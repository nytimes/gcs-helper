package handlers

import (
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/NYTimes/gcs-helper/v3/internal/testhelper"
	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/sirupsen/logrus"
)

func testMapServer(t *testing.T, cfg Config) (string, func()) {
	server, err := fakestorage.NewServerWithOptions(fakestorage.Options{
		InitialObjects: testhelper.FakeObjects,
		NoListener:     true,
	})
	if err != nil {
		t.Fatal(err)
	}
	httpServer := httptest.NewServer(Map(cfg, server.Client()))
	return httpServer.URL, func() {
		httpServer.Close()
		server.Stop()
	}
}

func testProxyServer(t *testing.T, cfg Config) (string, func()) {
	logger := logrus.New()
	logger.Out = ioutil.Discard
	server, err := fakestorage.NewServerWithOptions(fakestorage.Options{
		InitialObjects: testhelper.FakeObjects,
		NoListener:     true,
	})
	if err != nil {
		t.Fatal(err)
	}
	httpServer := httptest.NewServer(Proxy(cfg, server.HTTPClient()))
	return httpServer.URL, func() {
		httpServer.Close()
		server.Stop()
	}
}
