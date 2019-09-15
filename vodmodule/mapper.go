package vodmodule

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/flavioribeiro/mediainfo"
	log "github.com/sirupsen/logrus"
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

	// Optional and specific to cbsinteractive case
	ChapterBreaks string

	// Optional used  to build url and fetch duration with mediainfo
	ProxyListen string

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
	if opts.ChapterBreaks != "" {
		r.Durations, _ = m.chapterBreaksToDurations(ctx, opts.ChapterBreaks, opts.ProxyListen, opts.ProxyEndpoint, opts.Prefix)
	}
	r.Sequences, err = m.getSequences(ctx, opts.Prefix, opts.Filter, opts.ProxyEndpoint, r.Durations)
	return r, err
}

func (m *Mapper) getSequences(ctx context.Context, prefix string, filter *regexp.Regexp, proxyEndpoint string, durations []int) ([]Sequence, error) {
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
				if len(durations) > 0 {
					previousDuration := 0
					clips := []Clip{}
					for i := range durations {
						clip := Clip{
							Type:     "source",
							Path:     proxyEndpoint + "/" + obj.Name,
							ClipFrom: previousDuration,
						}
						if i != len(durations)-1 {
							clip.ClipTo = durations[i] + previousDuration
						}
						clips = append(clips, clip)
						previousDuration = durations[i] + previousDuration
					}
					sequence := Sequence{Clips: clips}
					seqs = append(seqs, sequence)
				} else {
					sequence := Sequence{
						Clips: []Clip{
							{
								Type: "source",
								Path: proxyEndpoint + "/" + obj.Name,
							},
						},
					}
					seqs = append(seqs, sequence)
				}
			}
		}
		if err == iterator.Done {
			return seqs, nil
		}
	}
	return nil, err
}

func (m *Mapper) chapterBreaksToDurations(ctx context.Context, chapterBreaks string, proxyListen string, endpoint string, prefix string) ([]int, error) {
	var err error
	var obj *storage.ObjectAttrs

	previousTimestamp := 0
	totalDurations := 0
	splittedChapterBreaks := strings.Split(chapterBreaks, ",")
	result := make([]int, 0) // is there something better than this?
	for i := range splittedChapterBreaks {
		chapterBreakInMs := m.convertChapterBreakInMs(splittedChapterBreaks[i])
		result = append(result, chapterBreakInMs-previousTimestamp)
		totalDurations = totalDurations + chapterBreakInMs
		previousTimestamp = chapterBreakInMs
	}

	logger := log.New()
	iter := m.bucket.Objects(ctx, &storage.Query{
		Prefix:    prefix,
		Delimiter: "/",
	})
	obj, _ = iter.Next() // ignoring error for now
	fileURL := fmt.Sprintf("http://127.0.0.1%s%s/%s", proxyListen, endpoint, obj.Name)
	mi, _ := mediainfo.New(fileURL, logger, "sample_file")

	result = append(result, int(mi.General.Duration.Val)-totalDurations) // last piece should have all the content

	return result, err
}

// convertChapterBreakInMs is amazing
func (m *Mapper) convertChapterBreakInMs(chapterBreak string) int {
	var hrs, mins, secs int
	splittedChapter := strings.Split(chapterBreak, ":")
	if len(splittedChapter) == 2 {
		mins, _ = strconv.Atoi(splittedChapter[0])
		secs, _ = strconv.Atoi(splittedChapter[1])
	} else {
		hrs, _ = strconv.Atoi(splittedChapter[0])
		mins, _ = strconv.Atoi(splittedChapter[1])
		secs, _ = strconv.Atoi(splittedChapter[2])
	}
	return hrs*3600000 + mins*60000 + secs*1000
}
