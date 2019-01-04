package vodmodule

import (
	"context"
	"regexp"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/NYTimes/gcs-helper/v3/vodmodule/testhelper"
	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/google/go-cmp/cmp"
)

func TestMap(t *testing.T) {
	server, bucket := fakeBucketHandle("my-bucket")
	defer server.Stop()
	mapper := NewMapper(bucket)

	var tests = []struct {
		name     string
		input    MapOptions
		expected Mapping
	}{
		{
			"list of files, no filter",
			MapOptions{
				Prefix: "videos/video/",
			},
			Mapping{
				Sequences: []Sequence{
					{
						Clips: []Clip{
							{Type: "source", Path: "/my-bucket/videos/video/28043_1_video_1080p.mp4"},
						},
					},
					{
						Clips: []Clip{
							{Type: "source", Path: "/my-bucket/videos/video/77071_1_caption_wg_240p_001f8ea7-749b-4d43-7bd5-b357e4e24f32.srt"},
						},
					},
					{
						Clips: []Clip{
							{Type: "source", Path: "/my-bucket/videos/video/77071_1_caption_wg_240p_001f8ea7-749b-4d43-7bd5-b357e4e24f32.vtt"},
						},
					},
					{
						Clips: []Clip{
							{Type: "source", Path: "/my-bucket/videos/video/video1_480p.mp4"},
						},
					},
					{
						Clips: []Clip{
							{Type: "source", Path: "/my-bucket/videos/video/video1_720p.mp4"},
						},
					},
				},
			},
		},
		{
			"list of files with filter",
			MapOptions{
				Prefix: "videos/video/",
				Filter: regexp.MustCompile(`.mp4$`),
			},
			Mapping{
				Sequences: []Sequence{
					{
						Clips: []Clip{
							{Type: "source", Path: "/my-bucket/videos/video/28043_1_video_1080p.mp4"},
						},
					},
					{
						Clips: []Clip{
							{Type: "source", Path: "/my-bucket/videos/video/video1_480p.mp4"},
						},
					},
					{
						Clips: []Clip{
							{Type: "source", Path: "/my-bucket/videos/video/video1_720p.mp4"},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			m, err := mapper.Map(context.TODO(), test.input)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(m, test.expected); diff != "" {
				t.Errorf("wrong mapping returned\nwant %#v\ngot  %#v", test.expected, m)
			}
		})
	}
}

func fakeBucketHandle(bucketName string) (*fakestorage.Server, *storage.BucketHandle) {
	server := fakestorage.NewServer(testhelper.FakeObjects)
	return server, server.Client().Bucket(bucketName)
}
