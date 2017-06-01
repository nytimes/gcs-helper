package main

import (
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

// Config represents the gcs-helper configuration that is loaded from the
// environment.
type Config struct {
	Listen     string `default:":8080"`
	BucketName string `envconfig:"BUCKET_NAME" required:"true"`
	LogLevel   string `envconfig:"LOG_LEVEL" default:"debug"`
}

func (c Config) logger() *logrus.Logger {
	level, err := logrus.ParseLevel(c.LogLevel)
	if err != nil {
		level = logrus.DebugLevel
	}

	logger := logrus.New()
	logger.Out = os.Stderr
	logger.Level = level
	return logger
}

func loadConfig() (Config, error) {
	var c Config
	err := envconfig.Process("gcs_helper", &c)
	return c, err
}
