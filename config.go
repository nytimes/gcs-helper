package main

import (
	"github.com/kelseyhightower/envconfig"
)

// Config represents the gcs-helper configuration that is loaded from the
// environment.
type Config struct {
	Listen     string `default:":8080"`
	BucketName string `envconfig:"BUCKET_NAME" required:"true"`
}

func loadConfig() (Config, error) {
	var c Config
	err := envconfig.Process("gcs_helper", &c)
	return c, err
}
