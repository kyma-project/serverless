package chart

import (
	"fmt"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/strvals"
)

type ImageReplace func(string) *flagsBuilder

type FlagsBuilder interface {
	Build() (map[string]interface{}, error)
	WithControllerConfiguration(CPUUtilizationPercentage string, requeueDuration string, buildExecutorArgs string, maxSimultaneousJobs string, healthzLivenessTimeout string) *flagsBuilder
	WithDefaultPresetFlags(defaultBuildJobPreset string, defaultRuntimePodPreset string) *flagsBuilder
	WithOptionalDependencies(publisherURL string, traceCollectorURL string) *flagsBuilder
	WithRegistryAddresses(registryAddress string, serverAddress string) *flagsBuilder
	WithRegistryCredentials(username string, password string) *flagsBuilder
	WithRegistryEnableInternal(enableInternal bool) *flagsBuilder
	WithRegistryHttpSecret(httpSecret string) *flagsBuilder
	WithManagedByLabel(string) *flagsBuilder
	WithNodePort(nodePort int64) *flagsBuilder
	WithLogLevel(logLevel string) *flagsBuilder
	WithLogFormat(logFormat string) *flagsBuilder
	//TODO: remove this method when buildless is enabled by default
	WithChartPath(chartPath string) *flagsBuilder
	WithImageFunctionBuildfulController(image string) *flagsBuilder
	WithImageFunctionController(image string) *flagsBuilder
	WithImageFunctionBuildInit(image string) *flagsBuilder
	WithImageFunctionInit(image string) *flagsBuilder
	WithImageRegistryInit(image string) *flagsBuilder
	WithImageFunctionRuntimeNodejs20(image string) *flagsBuilder
	WithImageFunctionRuntimeNodejs22(image string) *flagsBuilder
	WithImageFunctionRuntimePython312(image string) *flagsBuilder
	WithImageKanikoExecutor(image string) *flagsBuilder
	WithImageRegistry(image string) *flagsBuilder
}

type flagsBuilder struct {
	flags map[string]interface{}
}

func NewFlagsBuilder() FlagsBuilder {
	return &flagsBuilder{
		flags: map[string]interface{}{},
	}
}

func (fb *flagsBuilder) Build() (map[string]interface{}, error) {
	flags := map[string]interface{}{}
	for key, value := range fb.flags {
		flag := fmt.Sprintf("%s=%v", key, value)
		err := strvals.ParseInto(flag, flags)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse %s flag", flag)
		}
	}
	return flags, nil
}

func (fb *flagsBuilder) WithControllerConfiguration(CPUUtilizationPercentage, requeueDuration, buildExecutorArgs, maxSimultaneousJobs, healthzLivenessTimeout string) *flagsBuilder {
	optionalFlags := []struct {
		key   string
		value string
	}{
		{"targetCPUUtilizationPercentage", CPUUtilizationPercentage},
		{"functionRequeueDuration", requeueDuration},
		{"functionBuildExecutorArgs", buildExecutorArgs},
		{"functionBuildMaxSimultaneousJobs", maxSimultaneousJobs},
		{"healthzLivenessTimeout", healthzLivenessTimeout},
	}

	for _, flag := range optionalFlags {
		if flag.value != "" {
			fullPath := fmt.Sprintf("containers.manager.configuration.data.%s", flag.key)
			fb.flags[fullPath] = flag.value
		}
	}

	return fb
}

func (fb *flagsBuilder) WithOptionalDependencies(publisherURL, traceCollectorURL string) *flagsBuilder {
	fb.flags["containers.manager.configuration.data.functionTraceCollectorEndpoint"] = traceCollectorURL
	fb.flags["containers.manager.configuration.data.functionPublisherProxyAddress"] = publisherURL
	return fb
}

func (fb *flagsBuilder) WithRegistryEnableInternal(enableInternal bool) *flagsBuilder {
	fb.flags["dockerRegistry.enableInternal"] = enableInternal
	return fb
}

func (fb *flagsBuilder) WithRegistryCredentials(username, password string) *flagsBuilder {
	fb.flags["dockerRegistry.username"] = username
	fb.flags["dockerRegistry.password"] = password
	return fb
}

func (fb *flagsBuilder) WithRegistryAddresses(registryAddress, serverAddress string) *flagsBuilder {
	fb.flags["dockerRegistry.registryAddress"] = registryAddress
	fb.flags["dockerRegistry.serverAddress"] = serverAddress

	return fb
}

func (fb *flagsBuilder) WithRegistryHttpSecret(httpSecret string) *flagsBuilder {
	fb.flags["docker-registry.rollme"] = "dontrollplease"
	fb.flags["docker-registry.registryHTTPSecret"] = httpSecret

	return fb
}

func (fb *flagsBuilder) WithDefaultPresetFlags(defaultBuildJobPreset, defaultRuntimePodPreset string) *flagsBuilder {
	if defaultRuntimePodPreset != "" {
		fb.flags["containers.manager.configuration.data.resourcesConfiguration.function.resources.defaultPreset"] = defaultRuntimePodPreset
	}

	if defaultBuildJobPreset != "" {
		fb.flags["containers.manager.configuration.data.resourcesConfiguration.buildJob.resources.defaultPreset"] = defaultBuildJobPreset
	}

	return fb
}

func (fb *flagsBuilder) WithManagedByLabel(managedBy string) *flagsBuilder {
	fb.flags["global.commonLabels.app\\.kubernetes\\.io/managed-by"] = managedBy
	return fb
}

func (fb *flagsBuilder) WithNodePort(nodePort int64) *flagsBuilder {
	fb.flags["global.registryNodePort"] = nodePort
	return fb
}

func (fb *flagsBuilder) WithLogLevel(logLevel string) *flagsBuilder {
	if logLevel != "" {
		fb.flags["containers.manager.logConfiguration.data.logLevel"] = logLevel
	}

	return fb
}

func (fb *flagsBuilder) WithLogFormat(logFormat string) *flagsBuilder {
	if logFormat != "" {
		fb.flags["containers.manager.logConfiguration.data.logFormat"] = logFormat
	}

	return fb
}

// TODO: remove this method when buildless is enabled by default
func (fb *flagsBuilder) WithChartPath(chartPath string) *flagsBuilder {
	fb.flags["chartPath"] = chartPath
	return fb
}

// temporary name until buildless takes over
func (fb *flagsBuilder) WithImageFunctionBuildfulController(image string) *flagsBuilder {
	fb.flags["global.images.function_buildful_controller"] = image
	return fb
}

func (fb *flagsBuilder) WithImageFunctionController(image string) *flagsBuilder {
	fb.flags["global.images.function_controller"] = image
	return fb
}

func (fb *flagsBuilder) WithImageFunctionBuildInit(image string) *flagsBuilder {
	fb.flags["global.images.function_build_init"] = image
	return fb
}

func (fb *flagsBuilder) WithImageFunctionInit(image string) *flagsBuilder {
	fb.flags["global.images.function_init"] = image
	return fb
}

func (fb *flagsBuilder) WithImageRegistryInit(image string) *flagsBuilder {
	fb.flags["global.images.registry_init"] = image
	return fb
}

func (fb *flagsBuilder) WithImageFunctionRuntimeNodejs20(image string) *flagsBuilder {
	fb.flags["global.images.function_runtime_nodejs20"] = image
	return fb
}

func (fb *flagsBuilder) WithImageFunctionRuntimeNodejs22(image string) *flagsBuilder {
	fb.flags["global.images.function_runtime_nodejs22"] = image
	return fb
}

func (fb *flagsBuilder) WithImageFunctionRuntimePython312(image string) *flagsBuilder {
	fb.flags["global.images.function_runtime_python312"] = image
	return fb
}

func (fb *flagsBuilder) WithImageKanikoExecutor(image string) *flagsBuilder {
	fb.flags["global.images.kaniko_executor"] = image
	return fb
}

func (fb *flagsBuilder) WithImageRegistry(image string) *flagsBuilder {
	fb.flags["global.images.registry"] = image
	return fb
}
