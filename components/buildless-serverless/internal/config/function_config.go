package config

import (
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"os"
	"path/filepath"
	"time"
)

type FunctionConfig struct {
	MetricsPort                     string         `yaml:"metricsPort"`
	LeaderElectionEnabled           bool           `yaml:"leaderElectionEnabled"`
	LeaderElectionID                string         `yaml:"leaderElectionID"`
	SecretMutatingWebhookPort       int            `yaml:"secretMutatingWebhookPort"`
	HealthzPort                     string         `yaml:"healthzPort"`
	Images                          ImagesConfig   `yaml:"images"`
	RequeueDuration                 time.Duration  `yaml:"requeueDuration"`
	FunctionReadyRequeueDuration    time.Duration  `yaml:"functionReadyRequeueDuration"`
	PackageRegistryConfigSecretName string         `yaml:"packageRegistryConfigSecretName"`
	FunctionTraceCollectorEndpoint  string         `yaml:"functionTraceCollectorEndpoint"`
	FunctionPublisherProxyAddress   string         `yaml:"functionPublisherProxyAddress"`
	ResourceConfig                  ResourceConfig `yaml:"resourcesConfiguration"`
	InternalEndpointPort            string         `yaml:"internalEndpointPort"`
}

func defaultFunctionConfig() FunctionConfig {
	return FunctionConfig{
		MetricsPort:                     ":8080",
		LeaderElectionEnabled:           false,
		LeaderElectionID:                "serverless-controller-leader-election-helper",
		SecretMutatingWebhookPort:       8443,
		HealthzPort:                     ":8090",
		RequeueDuration:                 time.Minute,
		FunctionReadyRequeueDuration:    time.Minute * 5,
		PackageRegistryConfigSecretName: "buildless-serverless-package-registry-config",
		FunctionPublisherProxyAddress:   "http://eventing-publisher-proxy.kyma-system.svc.cluster.local/publish",
		InternalEndpointPort:            ":12137",
	}
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

var _ envconfig.Unmarshaler = &ResourceConfig{}

func (rc *ResourceConfig) Unmarshal(input string) error {
	err := yaml.Unmarshal([]byte(input), rc)
	return err
}

type FunctionResourceConfig struct {
	Resources Resources `yaml:"resources"`
}

type Resources struct {
	DefaultPreset    string   `yaml:"defaultPreset"`
	MinRequestCPU    Quantity `yaml:"minRequestCPU"`
	MinRequestMemory Quantity `yaml:"minRequestMemory"`
	Presets          Preset   `yaml:"presets"`
}

type Preset map[string]Resource

type Resource struct {
	RequestCPU    Quantity `yaml:"requestCpu"`
	RequestMemory Quantity `yaml:"requestMemory"`
	LimitCPU      Quantity `yaml:"limitCpu"`
	LimitMemory   Quantity `yaml:"limitMemory"`
}

type Quantity struct {
	Quantity resource.Quantity
}

func (q *Quantity) UnmarshalYAML(unmarshal func(interface{}) error) error {
	quantity := ""
	err := unmarshal(&quantity)
	if err != nil {
		return errors.Wrap(err, "while unmarshalling quantity")
	}
	out, err := resource.ParseQuantity(quantity)
	if err != nil {
		return errors.Wrap(err, "while parsing quantity")
	}
	q.Quantity = out
	return nil
}

func LoadFunctionConfig(path string) (FunctionConfig, error) {
	cfg := defaultFunctionConfig()

	cleanPath := filepath.Clean(path)
	yamlFile, err := os.ReadFile(cleanPath)
	if err != nil {
		return cfg, err
	}

	err = yaml.Unmarshal(yamlFile, &cfg)
	return cfg, err
}

func (r Resource) ToResourceRequirements() corev1.ResourceRequirements {
	result := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    r.LimitCPU.Quantity,
			corev1.ResourceMemory: r.LimitMemory.Quantity,
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    r.RequestCPU.Quantity,
			corev1.ResourceMemory: r.RequestMemory.Quantity,
		},
	}
	return result
}
