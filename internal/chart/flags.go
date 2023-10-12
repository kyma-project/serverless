package chart

import "strings"

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
	dockerRegistryChartParams map[string]interface{}
	dockerRegistryParams      map[string]interface{}
	containersDataParams      map[string]interface{}
	webhookValuesParams       map[string]interface{}
	globalParams              map[string]interface{}
	flags                     map[string]interface{}
}

func NewFlagsBuilder() FlagsBuilder {
	return &flagsBuilder{
		dockerRegistryChartParams: map[string]interface{}{},
		dockerRegistryParams:      map[string]interface{}{},
		containersDataParams:      map[string]interface{}{},
		webhookValuesParams:       map[string]interface{}{},
		globalParams:              map[string]interface{}{},
		flags:                     map[string]interface{}{},
	}
}

func (fb *flagsBuilder) Build() map[string]interface{} {
	//
	//if paramsAreNotEmpty(fb.containersDataParams) {
	//	flags["containers"] = map[string]interface{}{
	//		"manager": map[string]interface{}{
	//			"configuration": map[string]interface{}{
	//				"data": fb.containersDataParams,
	//			},
	//		},
	//	}
	//}
	//
	//if paramsAreNotEmpty(fb.globalParams) {
	//	flags["global"] = fb.globalParams
	//}
	//
	//if paramsAreNotEmpty(fb.webhookValuesParams) {
	//	flags["webhook"] = map[string]interface{}{
	//		"values": fb.webhookValuesParams,
	//	}
	//}
	//
	//if paramsAreNotEmpty(fb.dockerRegistryParams) {
	//	flags["dockerRegistry"] = fb.dockerRegistryParams
	//}
	//
	//if paramsAreNotEmpty(fb.dockerRegistryChartParams) {
	//	flags["docker-registry"] = fb.dockerRegistryChartParams
	//}

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
	return fb.flags
}

func paramsAreNotEmpty(params map[string]interface{}) bool {
	return len(params) > 0
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
			fb.containersDataParams[flag.key] = flag.value
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
	fb.dockerRegistryParams["enableInternal"] = enableInternal

	return fb
}

func (fb *flagsBuilder) WithRegistryCredentials(username, password string) *flagsBuilder {
	fb.dockerRegistryParams["username"] = username
	fb.dockerRegistryParams["password"] = password

	return fb
}

func (fb *flagsBuilder) WithRegistryAddresses(registryAddress, serverAddress string) *flagsBuilder {
	fb.dockerRegistryParams["registryAddress"] = registryAddress
	fb.dockerRegistryParams["serverAddress"] = serverAddress

	return fb
}

func (fb *flagsBuilder) WithRegistryHttpSecret(httpSecret string) *flagsBuilder {
	fb.dockerRegistryChartParams["rollme"] = "dontrollplease"
	fb.dockerRegistryChartParams["registryHTTPSecret"] = httpSecret

	return fb
}

func (fb *flagsBuilder) WithDefaultPresetFlags(defaultBuildJobPreset, defaultRuntimePodPreset string) *flagsBuilder {
	if defaultRuntimePodPreset != "" {
		fb.webhookValuesParams["function"] = map[string]interface{}{
			"resources": map[string]interface{}{
				"defaultPreset": defaultRuntimePodPreset,
			},
		}
	}

	if defaultBuildJobPreset != "" {
		fb.webhookValuesParams["buildJob"] = map[string]interface{}{
			"resources": map[string]interface{}{
				"defaultPreset": defaultBuildJobPreset,
			},
		}
	}

	return fb
}

func (fb *flagsBuilder) WithNodePort(nodePort int64) *flagsBuilder {
	fb.globalParams["registryNodePort"] = nodePort

	return fb
}
