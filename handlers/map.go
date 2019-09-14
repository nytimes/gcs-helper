package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/NYTimes/gcs-helper/v3/vodmodule"
)

// Map returns the map handler.
func Map(c Config, client *storage.Client) http.Handler {
	mapper := vodmodule.NewMapper(client.Bucket(c.BucketName))
	filter := regexp.MustCompile(c.Map.RegexFilter)
	logger := c.Logger()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		fmt.Println("URL:", r.URL)
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		chapterBreaks := ""
		if val, ok := r.URL.Query()["breaks"]; ok {
			if val[0] != "_" {
				chapterBreaks = val[0]
			}
		}
		prefix := strings.TrimLeft(r.URL.Path, "/")
		if prefix == "" {
			http.Error(w, "prefix cannot be empty", http.StatusBadRequest)
			return
		}
		m, err := mapper.Map(r.Context(), vodmodule.MapOptions{
			Prefix:        prefix,
			Filter:        filter,
			ProxyEndpoint: c.Proxy.Endpoint,
			ChapterBreaks: chapterBreaks,
		})
		if err != nil {
			logger.WithError(err).WithField("prefix", prefix).Error("failed to map request")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m)
	})
}
