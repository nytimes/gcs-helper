package main

import (
	"context"
	"encoding/json"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
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
	bucketHandle := client.Bucket(c.BucketName)
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
		m, err := getPrefixMapping(prefix, c, bucketHandle)
		if err != nil && err != iterator.Done {
			logger.WithError(err).WithField("prefix", prefix).Error("failed to map request")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m)
	}
}

func getPrefixMapping(prefix string, config Config, bucketHandle *storage.BucketHandle) (mapping, error) {
	m := mapping{Sequences: []sequence{}}
	for _, p := range getPrefixes(prefix, config) {
		sequences, err := expandPrefix(p, config, bucketHandle)
		if err != nil {
			return m, err
		}
		m.Sequences = append(m.Sequences, sequences...)
	}
	return m, nil
}

func getPrefixes(originalPrefix string, config Config) []string {
	prefixes := []string{originalPrefix}
	_, lastPart := path.Split(originalPrefix)
	for _, p := range config.MapExtraPrefixes {
		prefixes = append(prefixes, path.Join(p, lastPart))
	}
	return prefixes
}

func expandPrefix(prefix string, config Config, bucketHandle *storage.BucketHandle) ([]sequence, error) {
	var err error
	for i := 0; i < maxTry; i++ {
		iter := bucketHandle.Objects(context.Background(), &storage.Query{
			Prefix:    prefix,
			Delimiter: "/",
		})
		var obj *storage.ObjectAttrs
		sequences := []sequence{}
		obj, err = iter.Next()
		for ; err == nil; obj, err = iter.Next() {
			ext := filepath.Ext(obj.Name)
			if config.checkExtension(ext) {
				sequences = append(sequences, sequence{
					Clips: []clip{{Type: "source", Path: "/" + obj.Name}},
				})
			}
		}
		if err == iterator.Done {
			return sequences, nil
		}
	}
	return nil, err
}
