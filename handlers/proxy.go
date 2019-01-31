package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"github.com/sirupsen/logrus"
)

type codeWrapper struct {
	code int
	http.ResponseWriter
}

func (w *codeWrapper) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

type proxyHandler struct {
	config Config
	logger *logrus.Logger
	hc     *http.Client
}

func (h *proxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	resp := codeWrapper{ResponseWriter: w}
	var err error

	defer r.Body.Close()
	defer func() {
		if err != nil || h.logger.Level >= logrus.DebugLevel {
			fields := logrus.Fields{
				"method":        r.Method,
				"ellapsed":      time.Since(start).String(),
				"path":          r.URL.RequestURI(),
				"proxyEndpoint": h.config.Proxy.Endpoint,
				"response":      resp.code,
			}
			for _, header := range h.config.Proxy.LogHeaders {
				if value := r.Header.Get(header); value != "" {
					fields["ReqHeader/"+header] = value
				}
			}
			entry := h.logger.WithFields(fields)
			if err != nil {
				entry.WithError(err).Error("failed to handle request")
			} else {
				entry.Debug("finished handling request")
			}
		}
	}()

	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path == "/" {
		w.WriteHeader(http.StatusOK)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), h.config.Proxy.Timeout)
	defer cancel()

	host := "storage.googleapis.com"
	if !h.config.Proxy.BucketOnPath {
		host = h.config.BucketName + "." + host
	}
	url := fmt.Sprintf("https://%s%s", host, r.URL.RequestURI())
	// no support for request body, do we care? :)
	gcsReq, err := http.NewRequest(r.Method, url, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	gcsReq = gcsReq.WithContext(ctx)
	for name, values := range r.Header {
		for _, value := range values {
			gcsReq.Header.Add(name, value)
		}
	}
	gcsResp, err := h.hc.Do(gcsReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer gcsResp.Body.Close()

	for name, values := range gcsResp.Header {
		for _, value := range values {
			resp.Header().Add(name, value)
		}
	}
	resp.WriteHeader(gcsResp.StatusCode)
	io.Copy(resp, gcsResp.Body)
}

// Proxy returns the proxy handler.
func Proxy(c Config, client *storage.Client) http.Handler {
	logger := c.Logger()
	hc, err := c.Client.HTTPClient()
	if err != nil {
		logger.Fatalf("failed to initialize http client: %v", err.Error())
	}
	return &proxyHandler{logger: logger, hc: hc, config: c}
}
