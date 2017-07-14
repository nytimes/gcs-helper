package main

import (
	"os"
	"reflect"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestLoadConfig(t *testing.T) {
	setEnvs(map[string]string{
		"GCS_HELPER_LISTEN":            "0.0.0.0:3030",
		"GCS_HELPER_BUCKET_NAME":       "some-bucket",
		"GCS_HELPER_LOG_LEVEL":         "info",
		"GCS_HELPER_MAP_PREFIX":        "/map/",
		"GCS_HELPER_PROXY_PREFIX":      "/proxy/",
		"GCS_HELPER_PROXY_LOG_HEADERS": "Accept,Range",
		"GCS_HELPER_MAP_EXTENSIONS":    ".mp4,.vtt,.srt",
	})
	config, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	expectedConfig := Config{
		BucketName:      "some-bucket",
		Listen:          "0.0.0.0:3030",
		LogLevel:        "info",
		MapPrefix:       "/map/",
		ProxyPrefix:     "/proxy/",
		MapExtensions:   []string{".mp4", ".vtt", ".srt"},
		ProxyLogHeaders: []string{"Accept", "Range"},
	}
	if !reflect.DeepEqual(config, expectedConfig) {
		t.Errorf("wrong config returned\nwant %#v\ngot  %#v", expectedConfig, config)
	}
}

func TestLoadConfigDefaultValues(t *testing.T) {
	setEnvs(map[string]string{"GCS_HELPER_BUCKET_NAME": "some-bucket"})
	config, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	expectedConfig := Config{
		BucketName: "some-bucket",
		Listen:     ":8080",
		LogLevel:   "debug",
	}
	if !reflect.DeepEqual(config, expectedConfig) {
		t.Errorf("wrong config returned\nwant %#v\ngot  %#v", expectedConfig, config)
	}
}

func TestConfigLogger(t *testing.T) {
	setEnvs(map[string]string{"GCS_HELPER_BUCKET_NAME": "some-bucket", "GCS_HELPER_LOG_LEVEL": "info"})
	config, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	logger := config.logger()
	if logger.Out != os.Stdout {
		t.Errorf("wrong log output, want os.Stdout, got %#v", logger.Out)
	}
	if logger.Level != logrus.InfoLevel {
		t.Errorf("wrong log leve, want InfoLevel (%v), got %v", logrus.InfoLevel, logger.Level)
	}
}

func TestConfigLoggerInvalidLevel(t *testing.T) {
	setEnvs(map[string]string{"GCS_HELPER_BUCKET_NAME": "some-bucket", "GCS_HELPER_LOG_LEVEL": "dunno"})
	config, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	logger := config.logger()
	if logger.Out != os.Stdout {
		t.Errorf("wrong log output, want os.Stdout, got %#v", logger.Out)
	}
	if logger.Level != logrus.DebugLevel {
		t.Errorf("wrong log leve, want DebugLevel (%v), got %v", logrus.DebugLevel, logger.Level)
	}
}

func TestConfigCheckExtension(t *testing.T) {
	config := Config{
		MapExtensions: []string{".vtt", ".mp4"},
	}
	var tests = []struct {
		input  string
		output bool
	}{
		{".vtt", true},
		{".mp4", true},
		{".mp", false},
		{"mp4", false},
		{"vtt", false},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			r := config.checkExtension(test.input)
			if r != test.output {
				t.Errorf("config.checkExtension(%q): expected %v, got %v", test.input, test.output, r)
			}
		})
	}
}

func TestLoadConfigValidation(t *testing.T) {
	setEnvs(nil)
	config, err := loadConfig()
	if err == nil {
		t.Error("unexpected <nil> error")
	}
	expectedConfig := Config{Listen: ":8080"}
	if !reflect.DeepEqual(config, expectedConfig) {
		t.Errorf("wrong config returned\nwant %#v\ngot  %#v", expectedConfig, config)
	}
}

func setEnvs(envs map[string]string) {
	os.Clearenv()
	for name, value := range envs {
		os.Setenv(name, value)
	}
}
