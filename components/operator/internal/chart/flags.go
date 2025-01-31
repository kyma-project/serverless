package chart

import (
	"fmt"
	"strings"
)

type FlagsBuilder interface {
	Build() map[string]interface{}
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
		flagPath := strings.FieldsFunc(key, fieldsFuncWithExtrudes(key))
		flagPath = removeExcludesFromFlagPath(flagPath)
		appendFlag(flags, flagPath, value)
	}
	return flags
}

func fieldsFuncWithExtrudes(flag string) func(rune) bool {
	index := 0
	return func(r rune) bool {
		split := shouldBeSplit(flag, index, r)

		// increase index to know on which rune we are right now
		index++

		return split
	}
}

func shouldBeSplit(flag string, index int, r rune) bool {
	if r != '.' {
		return false
	}

	return index > 0 && flag[index-1] != '\\'
}

func removeExcludesFromFlagPath(flagPath []string) []string {
	for i := range flagPath {
		flagPath[i] = strings.ReplaceAll(flagPath[i], "\\", "")
	}
	return flagPath
}

func appendFlag(flags map[string]interface{}, flagPath []string, value interface{}) {
	currentFlag := flags
	for i, pathPart := range flagPath {
		createIfEmpty(currentFlag, pathPart)
		if lastElement(flagPath, i) {
			currentFlag[pathPart] = value
		} else {
			currentFlag = nextDeeperFlag(currentFlag, pathPart)
		}
	}
}

func createIfEmpty(flags map[string]interface{}, key string) {
	if _, ok := flags[key]; !ok {
		flags[key] = map[string]interface{}{}
	}
}

func lastElement(values []string, i int) bool {
	return i == len(values)-1
}

func nextDeeperFlag(currentFlag map[string]interface{}, path string) map[string]interface{} {
	return currentFlag[path].(map[string]interface{})
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
