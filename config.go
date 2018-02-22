package main

import (
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

// Config represents the gcs-helper configuration that is loaded from the
// environment.
type Config struct {
	Listen              string        `default:":8080"`
	BucketName          string        `envconfig:"BUCKET_NAME" required:"true"`
	LogLevel            string        `envconfig:"LOG_LEVEL" default:"debug"`
	ProxyLogHeaders     []string      `envconfig:"PROXY_LOG_HEADERS"`
	ProxyPrefix         string        `envconfig:"PROXY_PREFIX"`
	ProxyTimeout        time.Duration `envconfig:"PROXY_TIMEOUT" default:"10s"`
	MapPrefix           string        `envconfig:"MAP_PREFIX"`
	ExtraResourcesToken string        `envconfig:"EXTRA_RESOURCES_TOKEN"`
	MapRegexFilter      string        `envconfig:"MAP_REGEX_FILTER"`
	MapRegexHDFilter    string        `envconfig:"MAP_REGEX_HD_FILTER"`
	MapExtraPrefixes    []string      `envconfig:"MAP_EXTRA_PREFIXES"`
	MapExtensionSplit   bool          `envconfig:"MAP_EXTENSION_SPLIT"`
	ProxyBucketOnPath   bool          `envconfig:"PROXY_BUCKET_ON_PATH"`
	ClientConfig        ClientConfig
}

// ClientConfig contains configuration for the GCS client communication.
//
// It contains options related to timeouts and keep-alive connections.
type ClientConfig struct {
	Timeout         time.Duration `envconfig:"GCS_CLIENT_TIMEOUT" default:"2s"`
	IdleConnTimeout time.Duration `envconfig:"GCS_CLIENT_IDLE_CONN_TIMEOUT" default:"120s"`
	MaxIdleConns    int           `envconfig:"GCS_CLIENT_MAX_IDLE_CONNS" default:"10"`
}

func (c Config) logger() *logrus.Logger {
	level, err := logrus.ParseLevel(c.LogLevel)
	if err != nil {
		level = logrus.DebugLevel
	}

	logger := logrus.New()
	logger.Out = os.Stdout
	logger.Level = level
	return logger
}

func loadConfig() (Config, error) {
	var c Config
	err := envconfig.Process("gcs_helper", &c)
	return c, err
}
