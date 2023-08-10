package chart

func AppendContainersFlags(flags map[string]interface{}, publisherURL, traceCollectorURL, CPUUtilizationPercentage, requeueDuration, buildExecutorArgs, maxSimultaneousJobs, healthzLivenessTimeout, requestBodyLimitMb, timeoutSec string) map[string]interface{} {
	flags["containers"] = map[string]interface{}{
		"manager": map[string]interface{}{
			"envs": map[string]interface{}{
				"functionTraceCollectorEndpoint": map[string]interface{}{
					"value": traceCollectorURL,
				},
				"functionPublisherProxyAddress": map[string]interface{}{
					"value": publisherURL,
				},
				"targetCPUUtilizationPercentage":   getValueOrEmpty(CPUUtilizationPercentage),
				"functionRequeueDuration":          getValueOrEmpty(requeueDuration),
				"functionBuildExecutorArgs":        getValueOrEmpty(buildExecutorArgs),
				"functionBuildMaxSimultaneousJobs": getValueOrEmpty(maxSimultaneousJobs),
				"healthzLivenessTimeout":           getValueOrEmpty(healthzLivenessTimeout),
				"functionRequestBodyLimitMb":       getValueOrEmpty(requestBodyLimitMb),
				"functionTimeoutSec":               getValueOrEmpty(timeoutSec),
			},
		},
	}

	return flags
}

func getValueOrEmpty(value string) map[string]interface{} {
	if value == "" {
		return map[string]interface{}{}
	}
	return map[string]interface{}{"value": value}
}

/*
AppendNodePortFlag
nodePort must be int64, because when we compare old Flags with new flags, by default all integers are int64
*/
func AppendNodePortFlag(flags map[string]interface{}, nodePort int64) map[string]interface{} {
	flags["global"] = map[string]interface{}{
		"registryNodePort": nodePort,
	}
	return flags
}

func AppendInternalRegistryFlags(flags map[string]interface{}, enableInternal bool) map[string]interface{} {
	flags["dockerRegistry"] = map[string]interface{}{
		"enableInternal": enableInternal,
	}

	return flags
}

func AppendK3dRegistryFlags(flags map[string]interface{}, enableInternal bool, registryAddress, serverAddress string) map[string]interface{} {
	flags["dockerRegistry"] = map[string]interface{}{
		"enableInternal":  enableInternal,
		"registryAddress": registryAddress,
		"serverAddress":   serverAddress,
	}

	return flags
}

func AppendExternalRegistryFlags(flags map[string]interface{}, enableInternal bool, username, password, registryAddress, serverAddress string) map[string]interface{} {
	flags["dockerRegistry"] = map[string]interface{}{
		"enableInternal":  enableInternal,
		"username":        username,
		"password":        password,
		"registryAddress": registryAddress,
		"serverAddress":   serverAddress,
	}

	return flags
}

func EmptyFlags() map[string]interface{} {
	return map[string]interface{}{}
}
