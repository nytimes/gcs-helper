package main

import (
	"net/http"
	"strings"

	"cloud.google.com/go/storage"
)

func getHandler(c Config, client *storage.Client) http.HandlerFunc {
	proxyHandler := getProxyHandler(c, client)
	mapHandler := getMapHandler(c, client)

	return func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, c.Proxy.Endpoint):
			r.URL.Path = strings.Replace(r.URL.Path, c.Proxy.Endpoint, "", 1)
			proxyHandler(w, r)
		case strings.HasPrefix(r.URL.Path, c.Map.Endpoint):
			r.URL.Path = strings.Replace(r.URL.Path, c.Map.Endpoint, "", 1)
			mapHandler(w, r)
		case r.URL.Path == "/":
			// healthcheck
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}
}
