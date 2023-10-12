package chart

import (
	"fmt"
	"strings"
)

type FlagsBuilder interface {
	Build() map[string]interface{}
	WithControllerConfiguration(CPUUtilizationPercentage string, requeueDuration string, buildExecutorArgs string, maxSimultaneousJobs string, healthzLivenessTimeout string, requestBodyLimitMb string, timeoutSec string) *flagsBuilder
	WithDefaultPresetFlags(defaultBuildJobPreset string, defaultRuntimePodPreset string) *flagsBuilder
	WithOptionalDependencies(publisherURL string, traceCollectorURL string) *flagsBuilder
	WithRegistryAddresses(registryAddress string, serverAddress string) *flagsBuilder
	WithRegistryCredentials(username string, password string) *flagsBuilder
	WithRegistryEnableInternal(enableInternal bool) *flagsBuilder
	WithRegistryHttpSecret(httpSecret string) *flagsBuilder
	WithNodePort(nodePort int64) *flagsBuilder
}

type flagsBuilder struct {
	flags map[string]interface{}
}

func NewFlagsBuilder() FlagsBuilder {
	return &flagsBuilder{
		flags: map[string]interface{}{},
	}
}

func (fb *flagsBuilder) Build() map[string]interface{} {
	flags := map[string]interface{}{}
	for key, value := range fb.flags {
		valuePath := strings.Split(key, ".")
		currentPath := flags
		for i, path := range valuePath {
			if elem, ok := currentPath[path]; !ok {
				elem = map[string]interface{}{}
				currentPath[path] = elem
			}
			if i == len(valuePath)-1 {
				currentPath[path] = value
			} else {
				currentPath = currentPath[path].(map[string]interface{})
			}
		}
	}
	return flags
}

func (fb *flagsBuilder) WithControllerConfiguration(CPUUtilizationPercentage, requeueDuration, buildExecutorArgs, maxSimultaneousJobs, healthzLivenessTimeout, requestBodyLimitMb, timeoutSec string) *flagsBuilder {
	optionalFlags := []struct {
		key   string
		value string
	}{
		{"targetCPUUtilizationPercentage", CPUUtilizationPercentage},
		{"functionRequeueDuration", requeueDuration},
		{"functionBuildExecutorArgs", buildExecutorArgs},
		{"functionBuildMaxSimultaneousJobs", maxSimultaneousJobs},
		{"healthzLivenessTimeout", healthzLivenessTimeout},
		{"functionRequestBodyLimitMb", requestBodyLimitMb},
		{"functionTimeoutSec", timeoutSec},
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
		fb.flags["webhook.values.function.resources.defaultPreset"] = defaultRuntimePodPreset
	}

	if defaultBuildJobPreset != "" {
		fb.flags["webhook.values.buildJob.resources.defaultPreset"] = defaultBuildJobPreset
	}

	return fb
}

func (fb *flagsBuilder) WithNodePort(nodePort int64) *flagsBuilder {
	fb.flags["global.registryNodePort"] = nodePort
	return fb
}
