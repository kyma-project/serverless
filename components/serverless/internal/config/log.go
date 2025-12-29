package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	LogLevel  string `yaml:"logLevel"`
	LogFormat string `yaml:"logFormat"`
}

// LoadLogConfig - return cfg struct based on given path
func LoadLogConfig(path string) (Config, error) {
	cfg := Config{}

	cleanPath := filepath.Clean(path)
	yamlFile, err := os.ReadFile(cleanPath)
	if err != nil {
		return cfg, err
	}

	err = yaml.Unmarshal(yamlFile, &cfg)
	return cfg, err
}
