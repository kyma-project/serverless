package config

import (
	"time"
)

type FunctionConfig struct {
	ImageNodeJs20                   string
	ImageNodeJs22                   string
	ImagePython312                  string
	RequeueDuration                 time.Duration  `envconfig:"default=1m"`
	FunctionReadyRequeueDuration    time.Duration  `envconfig:"default=5m"`
	PackageRegistryConfigSecretName string         `envconfig:"default=buildless-serverless-package-registry-config"`
	FunctionTraceCollectorEndpoint  string         `envconfig:"optional"`
	FunctionPublisherProxyAddress   string         `envconfig:"default=http://eventing-publisher-proxy.kyma-system.svc.cluster.local/publish"`
	ResourceConfig                  ResourceConfig `envconfig:"optional"`
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
