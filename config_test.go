package main

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestLoadConfig(t *testing.T) {
	setEnvs(map[string]string{
		"GCS_HELPER_LISTEN":               "0.0.0.0:3030",
		"GCS_HELPER_BUCKET_NAME":          "some-bucket",
		"GCS_HELPER_LOG_LEVEL":            "info",
		"GCS_HELPER_MAP_PREFIX":           "/map/",
		"GCS_HELPER_PROXY_PREFIX":         "/proxy/",
		"GCS_HELPER_PROXY_LOG_HEADERS":    "Accept,Range",
		"GCS_HELPER_PROXY_TIMEOUT":        "20s",
		"GCS_HELPER_PROXY_BUCKET_ON_PATH": "true",
		"GCS_HELPER_MAP_REGEX_FILTERS":    `:(240|360|424|480|720|1080)p(\.mp4|[a-z0-9_-]{37}\.(vtt|srt))$,__HD:((720|1080)p\.mp4)|(\.(vtt|srt))$`,
		"GCS_HELPER_MAP_EXTRA_PREFIXES":   "subtitles/,mp4s/",
		"GCS_HELPER_MAP_EXTENSION_SPLIT":  "true",
		"GCS_CLIENT_TIMEOUT":              "60s",
		"GCS_CLIENT_IDLE_CONN_TIMEOUT":    "3m",
		"GCS_CLIENT_MAX_IDLE_CONNS":       "16",
	})
	config, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	expectedConfig := Config{
		BucketName: "some-bucket",
		Listen:     "0.0.0.0:3030",
		LogLevel:   "info",
		Proxy: ProxyConfig{
			Endpoint:     "/proxy/",
			LogHeaders:   []string{"Accept", "Range"},
			Timeout:      20 * time.Second,
			BucketOnPath: true,
		},
		Map: MapConfig{
			Endpoint:      "/map/",
			ExtraPrefixes: []string{"subtitles/", "mp4s/"},
			RegexFilters: map[string]string{
				"":     `(240|360|424|480|720|1080)p(\.mp4|[a-z0-9_-]{37}\.(vtt|srt))$`,
				"__HD": `((720|1080)p\.mp4)|(\.(vtt|srt))$`,
			},
			ExtensionSplit: true,
		},
		Client: ClientConfig{
			IdleConnTimeout: 3 * time.Minute,
			MaxIdleConns:    16,
			Timeout:         time.Minute,
		},
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
		Proxy: ProxyConfig{
			Timeout: 10 * time.Second,
		},
		Client: ClientConfig{
			IdleConnTimeout: 120 * time.Second,
			MaxIdleConns:    10,
			Timeout:         2 * time.Second,
		},
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
