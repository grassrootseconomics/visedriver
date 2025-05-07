//go:build !configreport

package config

import (
	"github.com/grassrootseconomics/go-vise/logging"
)

type Config struct{}

func NewConfig(logger logging.Vanilla) *Config {
	return &Config{}
}

func (c *Config) AddKey(key string, defaultValue string, sensitive bool, validator func(string) error) {
}

func (c *Config) GetValue(key string) (string, error) {
	return "", nil
}

func (c *Config) Report(level string) {}
