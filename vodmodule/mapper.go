package vodmodule

import (
	"context"
	"path"
	"regexp"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

const maxTries = 5

// Mapper provides the ability of mapping objects on a GCS bucket in the format
// expected by nginx-vod-module.
type Mapper struct {
	bucket *storage.BucketHandle
}

// NewMapper returns a mapper that will map content for prefix in the given
// BucketHandle.
func NewMapper(bucket *storage.BucketHandle) *Mapper {
	return &Mapper{bucket: bucket}
}

// MapOptions represents the set of options that can be passed to Map.
type MapOptions struct {
	ProxyEndpoint string
	Prefix        string

	// Optional regexp that is used to filter the list of objects.
	Filter *regexp.Regexp
}

// Map returns a Mapping object with the list of objects that match the given
// prefix. It supports a regular expression that is used to further filter (for
// example, if the caller only wants to return objects that with the ``.mp4``
// extension).
func (m *Mapper) Map(ctx context.Context, opts MapOptions) (Mapping, error) {
	var err error
	r := Mapping{}
	r.Sequences, err = m.getSequences(ctx, opts.Prefix, opts.Filter, opts.ProxyEndpoint)
	r.Durations = []int{20000, 5000, 6000000}
	return r, err
}

func (m *Mapper) getSequences(ctx context.Context, prefix string, filter *regexp.Regexp, proxyEndpoint string) ([]Sequence, error) {
	var err error
	for i := 0; i < maxTries; i++ {
		iter := m.bucket.Objects(ctx, &storage.Query{
			Prefix:    prefix,
			Delimiter: "/",
		})
		seqs := []Sequence{}
		var obj *storage.ObjectAttrs
		obj, err = iter.Next()
		for ; err == nil; obj, err = iter.Next() {
			filename := path.Base(obj.Name)
			if filter == nil || filter.MatchString(filename) {
				sequence := Sequence{
					Clips: []Clip{
						{
							Type:   "source",
							Path:   proxyEndpoint + "/" + obj.Name,
							ClipTo: 20000,
						},
						{
							Type:     "source",
							Path:     proxyEndpoint + "/" + obj.Name,
							ClipFrom: 10000,
							ClipTo:   15000,
						},
						{
							Type:     "source",
							Path:     proxyEndpoint + "/" + obj.Name,
							ClipFrom: 15000,
						},
					},
				}
				seqs = append(seqs, sequence)
			}
		}
		if err == iterator.Done {
			return seqs, nil
		}
	}
	return nil, err
}
