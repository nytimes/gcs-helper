package handlers

import (
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

// Config represents the gcs-helper configuration that is loaded from the
// environment.
type Config struct {
	Listen     string `default:":8080"`
	BucketName string `envconfig:"BUCKET_NAME" required:"true"`
	LogLevel   string `envconfig:"LOG_LEVEL" default:"debug"`
	Client     ClientConfig
	Map        MapConfig
	Proxy      ProxyConfig
}

// MapConfig contains configuration for the map mode.
type MapConfig struct {
	Endpoint    string `envconfig:"GCS_HELPER_MAP_PREFIX"`
	RegexFilter string `envconfig:"GCS_HELPER_MAP_REGEX_FILTER"`
}

// ProxyConfig contains configuration for the proxy mode.
type ProxyConfig struct {
	Endpoint     string        `envconfig:"GCS_HELPER_PROXY_PREFIX"`
	LogHeaders   []string      `envconfig:"GCS_HELPER_PROXY_LOG_HEADERS"`
	Timeout      time.Duration `envconfig:"GCS_HELPER_PROXY_TIMEOUT" default:"10s"`
	BucketOnPath bool          `envconfig:"GCS_HELPER_PROXY_BUCKET_ON_PATH"`
}

// ClientConfig contains configuration for the GCS client communication.
//
// It contains options related to timeouts and keep-alive connections.
type ClientConfig struct {
	Timeout         time.Duration `envconfig:"GCS_CLIENT_TIMEOUT" default:"2s"`
	IdleConnTimeout time.Duration `envconfig:"GCS_CLIENT_IDLE_CONN_TIMEOUT" default:"120s"`
	MaxIdleConns    int           `envconfig:"GCS_CLIENT_MAX_IDLE_CONNS" default:"10"`
}

func (c Config) Logger() *logrus.Logger {
	level, err := logrus.ParseLevel(c.LogLevel)
	if err != nil {
		level = logrus.DebugLevel
	}

	logger := logrus.New()
	logger.Out = os.Stdout
	logger.Level = level
	return logger
}

// LoadConfig loads the configuration from environment variables.
func LoadConfig() (Config, error) {
	var c Config
	err := envconfig.Process("gcs_helper", &c)
	return c, err
}
