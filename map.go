package main

import (
	"context"
	"encoding/json"
	"net/http"
	"path"
	"path/filepath"
	"regexp"
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
		var ext string
		prefix := strings.TrimLeft(r.URL.Path, "/")
		if c.Map.ExtensionSplit {
			ext = filepath.Ext(prefix)
			prefix = prefix[:len(prefix)-len(ext)]
		}
		if prefix == "" {
			http.Error(w, "prefix cannot be empty", http.StatusBadRequest)
			return
		}
		m, err := getPrefixMapping(prefix, ext, c, bucketHandle)
		if err != nil && err != iterator.Done {
			logger.WithError(err).WithField("prefix", prefix).Error("failed to map request")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		m = appendExtraResources(r, c, m)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m)
	}
}

func appendExtraResources(r *http.Request, config Config, m mapping) mapping {
	resources := r.URL.Query().Get(config.Map.ExtraResourcesToken)
	for _, resource := range strings.Split(resources, ",") {
		if resource != "" {
			m.Sequences = append(m.Sequences, sequence{
				Clips: []clip{{Type: "source", Path: resource}},
			})
		}
	}
	return m
}

func getPrefixMapping(prefix, ext string, config Config, bucketHandle *storage.BucketHandle) (mapping, error) {
	m := mapping{Sequences: []sequence{}}
	for _, p := range getPrefixes(prefix, config) {
		sequences, err := expandPrefix(p, ext, config, bucketHandle)
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
	for _, p := range config.Map.ExtraPrefixes {
		prefixes = append(prefixes, path.Join(p, lastPart))
	}
	return prefixes
}

func expandPrefix(prefix, ext string, config Config, bucketHandle *storage.BucketHandle) ([]sequence, error) {
	var err error
	var filterRegex string
	if strings.Contains(prefix, "__HD") {
		filterRegex = config.Map.RegexHDFilter
		prefix = strings.Replace(prefix, "__HD", "", 1)
	} else if ext != "" {
		filterRegex = regexp.QuoteMeta(ext) + "$"
	} else {
		filterRegex = config.Map.RegexFilter
	}
	for i := 0; i < maxTry; i++ {
		iter := bucketHandle.Objects(context.Background(), &storage.Query{
			Prefix:    prefix,
			Delimiter: "/",
		})
		var obj *storage.ObjectAttrs
		sequences := []sequence{}
		obj, err = iter.Next()
		for ; err == nil; obj, err = iter.Next() {
			filename := path.Base(obj.Name)
			matched, _ := regexp.MatchString(filterRegex, filename)
			if matched {
				sequences = append(sequences, sequence{
					Clips: []clip{{Type: "source", Path: "/" + obj.Bucket + "/" + obj.Name}},
				})
			}
		}
		if err == iterator.Done {
			return sequences, nil
		}
	}
	return nil, err
}
