package chart

var(
	EmptyFlags = map[string]interface{}{}
)

func AppendContainersFlags(flags map[string]interface{}, publisherURL, traceCollectorURL string) map[string]interface{} {
	flags["containers"] = map[string]interface{}{
		"manager": map[string]interface{}{
			"envs": map[string]interface{}{
				"functionTraceCollectorEndpoint": map[string]interface{}{
					"value": traceCollectorURL,
				},
				"functionPublisherProxyAddress": map[string]interface{}{
					"value": publisherURL,
				},
			},
		},
	}

	return flags
}

func AppendInternalRegistryFlags(flags map[string]interface{}, enableInternal bool) map[string]interface{} {
	flags["dockerRegistry"] = map[string]interface{}{
		"enableInternal": enableInternal,
	}

	return flags
}

func ApendK3dRegistryFlags(flags map[string]interface{}, enableInternal bool, registryAddress, serverAddress string) map[string]interface{} {
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
