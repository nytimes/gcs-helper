package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/NYTimes/gcs-helper/v3/vodmodule"
)

type mapping struct {
	Sequences []sequence `json:"sequences"`
}

type sequence struct {
	Clips []clip `json:"clips"`
}

type clip struct {
	Type string `json:"type"`
	Path string `json:"path"`
}

func getMapHandler(c Config, client *storage.Client) http.HandlerFunc {
	mapper := vodmodule.NewMapper(client.Bucket(c.BucketName))
	filter := regexp.MustCompile(c.Map.RegexFilter)
	logger := c.logger()
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}
