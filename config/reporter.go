//go:build configreport

package config

import (
	"fmt"

	"git.defalsify.org/vise.git/logging"
	"git.grassecon.net/grassrootseconomics/visedriver/env"
)

// ConfigValue represents a configuration key-value pair
type ConfigValue struct {
	Key       string
	Default   string
	Validator func(string) error
	Sensitive bool
}

// Config handles configuration management and reporting
type Config struct {
	values map[string]ConfigValue
	logger logging.Vanilla
}

func NewConfig(logger logging.Vanilla) *Config {
	return &Config{
		values: make(map[string]ConfigValue),
		logger: logger,
	}
}

// AddKey registers a new configuration key with optional validation
func (c *Config) AddKey(key string, defaultValue string, sensitive bool, validator func(string) error) {
	c.values[key] = ConfigValue{
		Key:       key,
		Default:   defaultValue,
		Validator: validator,
		Sensitive: sensitive,
	}
}

// GetValue returns the value for a given key, applying environment override if present
func (c *Config) GetValue(key string) (string, error) {
	// Find config value by key
	var cv ConfigValue
	for _, v := range c.values {
		if v.Key == key {
			cv = v
			break
		}
	}

	if cv.Key == "" {
		return "", fmt.Errorf("configuration key not found: %s", key)
	}

	// Get value from environment or default
	value := env.GetEnv(key, cv.Default)

	// Validate if validator exists
	if cv.Validator != nil && cv.Validator(value) != nil {
		return "", fmt.Errorf("invalid value for key %s", key)
	}

	return value, nil
}

// Report outputs all configuration values at the specified log level
func (c *Config) Report(level string) {
	c.logger.Debugf("Configuration Report:")
	for _, cv := range c.values {
		value, err := c.GetValue(cv.Key)
		if err != nil {
			c.logger.Errorf("Error getting value for %s: %v", cv.Key, err)
			continue
		}

		if cv.Sensitive {
			value = "****"
		}

		switch level {
		case "DEBUG":
			c.logger.Debugf("config set", cv.Key, value)
		case "INFO":
			c.logger.Infof("config set", cv.Key, value)
		case "WARN":
			c.logger.Warnf("config set", cv.Key, value)
		case "ERROR":
			c.logger.Errorf("config set", cv.Key, value)
		default:
			c.logger.Infof("config set", cv.Key, value)
		}
	}
}
