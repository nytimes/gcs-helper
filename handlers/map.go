package handlers

import (
	"encoding/json"
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
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		prefix := strings.TrimLeft(r.URL.Path, "/")
		if prefix == "" {
			http.Error(w, "prefix cannot be empty", http.StatusBadRequest)
			return
		}
		m, err := mapper.Map(r.Context(), vodmodule.MapOptions{
			Prefix: prefix,
			Filter: filter,
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
