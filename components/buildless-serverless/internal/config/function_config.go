package config

import (
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"time"
)

type FunctionConfig struct {
	Images                          ImagesConfig   `yaml:"images"`
	RequeueDuration                 time.Duration  `yaml:"requeueDuration"`
	FunctionReadyRequeueDuration    time.Duration  `yaml:"functionReadyRequeueDuration"`
	PackageRegistryConfigSecretName string         `yaml:"packageRegistryConfigSecretName"`
	FunctionTraceCollectorEndpoint  string         `yaml:"functionTraceCollectorEndpoint"`
	FunctionPublisherProxyAddress   string         `yaml:"functionPublisherProxyAddress"`
	ResourceConfig                  ResourceConfig `yaml:"resourceConfig"`
}

var defaultFunctionConfig = FunctionConfig{
	RequeueDuration:                 time.Minute,
	FunctionReadyRequeueDuration:    time.Minute * 5,
	PackageRegistryConfigSecretName: "buildless-serverless-package-registry-config",
	FunctionPublisherProxyAddress:   "http://eventing-publisher-proxy.kyma-system.svc.cluster.local/publish",
}

type ImagesConfig struct {
	NodeJs20    string `yaml:"nodejs20"`
	NodeJs22    string `yaml:"nodejs22"`
	Python312   string `yaml:"python312"`
	RepoFetcher string `yaml:"repoFetcher"`
}

type ResourceConfig struct {
	Function FunctionResourceConfig `yaml:"function"`
}

type FunctionResourceConfig struct {
	Resources Resources `yaml:"resources"`
}

type Resources struct {
	DefaultPreset string `yaml:"defaultPreset"`
	//TODO: add other fields
}

func LoadFunctionConfig(path string) (FunctionConfig, error) {
	cfg := defaultFunctionConfig

	cleanPath := filepath.Clean(path)
	yamlFile, err := os.ReadFile(cleanPath)
	if err != nil {
		return cfg, err
	}

	err = yaml.Unmarshal(yamlFile, &cfg)
	return cfg, err
}
