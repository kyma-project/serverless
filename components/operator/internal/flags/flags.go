package flags

import (
	"fmt"

	"github.com/kyma-project/manager-toolkit/installation/chart"
)

type ImageReplace func(string) *Builder

type Builder struct {
	chart.FlagsBuilder
}

func NewBuilder() *Builder {
	return &Builder{
		FlagsBuilder: chart.NewFlagsBuilder(),
	}
}

func (b *Builder) WithControllerConfiguration(CPUUtilizationPercentage, requeueDuration, buildExecutorArgs, maxSimultaneousJobs, healthzLivenessTimeout string) *Builder {
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
			b.With(fullPath, flag.value)
		}
	}

	return b
}

func (b *Builder) WithOptionalDependencies(publisherURL, traceCollectorURL string) *Builder {
	b.With("containers.manager.configuration.data.functionTraceCollectorEndpoint", traceCollectorURL)
	b.With("containers.manager.configuration.data.functionPublisherProxyAddress", publisherURL)
	return b
}

func (b *Builder) WithRegistryEnableInternal(enableInternal bool) *Builder {
	b.With("dockerRegistry.enableInternal", enableInternal)
	return b
}

func (b *Builder) WithRegistryCredentials(username, password string) *Builder {
	b.With("dockerRegistry.username", username)
	b.With("dockerRegistry.password", password)
	return b
}

func (b *Builder) WithRegistryAddresses(registryAddress, serverAddress string) *Builder {
	b.With("dockerRegistry.registryAddress", registryAddress)
	b.With("dockerRegistry.serverAddress", serverAddress)

	return b
}

func (b *Builder) WithRegistryHttpSecret(httpSecret string) *Builder {
	b.With("docker-registry.rollme", "dontrollplease")
	b.With("docker-registry.registryHTTPSecret", httpSecret)

	return b
}

func (b *Builder) WithDefaultPresetFlags(defaultBuildJobPreset, defaultRuntimePodPreset string) *Builder {
	if defaultRuntimePodPreset != "" {
		b.With("containers.manager.configuration.data.resourcesConfiguration.function.resources.defaultPreset", defaultRuntimePodPreset)
	}

	if defaultBuildJobPreset != "" {
		b.With("containers.manager.configuration.data.resourcesConfiguration.buildJob.resources.defaultPreset", defaultBuildJobPreset)
	}

	return b
}

func (b *Builder) WithManagedByLabel(managedBy string) *Builder {
	b.With("global.commonLabels.app\\.kubernetes\\.io/managed-by", managedBy)
	return b
}

func (b *Builder) WithNodePort(nodePort int64) *Builder {
	b.With("global.registryNodePort", nodePort)
	return b
}

func (b *Builder) WithLogLevel(logLevel string) *Builder {
	if logLevel != "" {
		b.With("containers.manager.logConfiguration.data.logLevel", logLevel)
	}

	return b
}

func (b *Builder) WithLogFormat(logFormat string) *Builder {
	if logFormat != "" {
		b.With("containers.manager.logConfiguration.data.logFormat", logFormat)
	}

	return b
}

// TODO: remove this method when buildless is enabled by default
func (b *Builder) WithChartPath(chartPath string) *Builder {
	b.With("chartPath", chartPath)
	return b
}

// temporary name until buildless takes over
func (b *Builder) WithImageFunctionBuildfulController(image string) *Builder {
	b.With("global.images.function_buildful_controller", image)
	return b
}

func (b *Builder) WithImageFunctionController(image string) *Builder {
	b.With("global.images.function_controller", image)
	return b
}

func (b *Builder) WithImageFunctionBuildInit(image string) *Builder {
	b.With("global.images.function_build_init", image)
	return b
}

func (b *Builder) WithImageFunctionInit(image string) *Builder {
	b.With("global.images.function_init", image)
	return b
}

func (b *Builder) WithImageRegistryInit(image string) *Builder {
	b.With("global.images.registry_init", image)
	return b
}

func (b *Builder) WithImageFunctionRuntimeNodejs20(image string) *Builder {
	b.With("global.images.function_runtime_nodejs20", image)
	return b
}

func (b *Builder) WithImageFunctionRuntimeNodejs22(image string) *Builder {
	b.With("global.images.function_runtime_nodejs22", image)
	return b
}

func (b *Builder) WithImageFunctionRuntimePython312(image string) *Builder {
	b.With("global.images.function_runtime_python312", image)
	return b
}

func (b *Builder) WithImageKanikoExecutor(image string) *Builder {
	b.With("global.images.kaniko_executor", image)
	return b
}

func (b *Builder) WithImageRegistry(image string) *Builder {
	b.With("global.images.registry", image)
	return b
}
