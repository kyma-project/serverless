package config

import (
	"os"
	"path/filepath"

	"github.com/vrischmann/envconfig"
	"gopkg.in/yaml.v2"
)

type Config struct {
	ChartPath     string `envconfig:"default=/module-chart"`
	LogLevel      string `envconfig:"default=info" yaml:"logLevel"`
	LogFormat     string `envconfig:"default=json" yaml:"logFormat"`
	LogConfigPath string `envconfig:"optional"`
}

func GetConfig(prefix string) (Config, error) {
	cfg := Config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	return cfg, err
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
