package extauthz

import (
	"github.com/sirupsen/logrus"
)

var log logrus.FieldLogger
var config *Config

// Init initializes the package with a logger and config.
func Init(logger *logrus.Logger, c *Config) {
	log = logger.WithField("package", "extauthz")
	config = c
	log.Infof("Base config: %+v", config)
}

// GetConfig returns the current config.
func GetConfig() *Config {
	return config
}
