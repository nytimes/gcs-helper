package main

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
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
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if r.Method != "GET" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		prefix := strings.TrimLeft(r.URL.Path, "/")
		if prefix == "" {
			http.Error(w, "prefix cannot be empty", http.StatusBadRequest)
			return
		}
		iter := bucketHandle.Objects(context.Background(), &storage.Query{
			Prefix:    prefix,
			Delimiter: "/",
		})
		m := mapping{Sequences: []sequence{}}
		obj, err := iter.Next()
		for ; err == nil; obj, err = iter.Next() {
			ext := filepath.Ext(obj.Name)
			if c.checkExtension(ext) {
				m.Sequences = append(m.Sequences, sequence{
					Clips: []clip{{Type: "source", Path: "/" + obj.Name}},
				})
			}
		}
		if err != iterator.Done {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m)
	}
}
