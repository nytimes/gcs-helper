package main

import (
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

// Config represents the gcs-helper configuration that is loaded from the
// environment.
type Config struct {
	Listen          string   `default:":8080"`
	BucketName      string   `envconfig:"BUCKET_NAME" required:"true"`
	LogLevel        string   `envconfig:"LOG_LEVEL" default:"debug"`
	ProxyLogHeaders []string `envconfig:"PROXY_LOG_HEADERS"`
	ProxyPrefix     string   `envconfig:"PROXY_PREFIX"`
	MapPrefix       string   `envconfig:"MAP_PREFIX"`
	MapExtensions   []string `envconfig:"MAP_EXTENSIONS"`
}

func (c Config) checkExtension(ext string) bool {
	for _, e := range c.MapExtensions {
		if e == ext {
			return true
		}
	}
	return false
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
