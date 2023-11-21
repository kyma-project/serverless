package chart

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_flagsBuilder_Build(t *testing.T) {
	t.Run("build empty flags", func(t *testing.T) {
		flags := NewFlagsBuilder().Build()
		require.Equal(t, map[string]interface{}{}, flags)
	})

	t.Run("build flags", func(t *testing.T) {
		expectedFlags := map[string]interface{}{
			"containers": map[string]interface{}{
				"manager": map[string]interface{}{
					"configuration": map[string]interface{}{
						"data": map[string]interface{}{
							"functionBuildExecutorArgs":        "testBuildExecutorArgs",
							"functionBuildMaxSimultaneousJobs": "testMaxSimultaneousJobs",
							"functionPublisherProxyAddress":    "testPublisherURL",
							"functionRequestBodyLimitMb":       "testRequestBodyLimitMb",
							"functionRequeueDuration":          "testRequeueDuration",
							"functionTimeoutSec":               "testTimeoutSec",
							"functionTraceCollectorEndpoint":   "testCollectorURL",
							"healthzLivenessTimeout":           "testHealthzLivenessTimeout",
							"targetCPUUtilizationPercentage":   "testCPUUtilizationPercentage",
							"namespaceExcludedNames":           "testNamespace;testNamespace2",
						},
					},
				},
			},
			"docker-registry": map[string]interface{}{
				"registryHTTPSecret": "testHttpSecret",
				"rollme":             "dontrollplease",
			},
			"dockerRegistry": map[string]interface{}{
				"enableInternal":  false,
				"password":        "testPassword",
				"registryAddress": "testRegistryAddress",
				"serverAddress":   "testServerAddress",
				"username":        "testUsername"},
			"global": map[string]interface{}{
				"registryNodePort": int64(1234),
			},
			"webhook": map[string]interface{}{
				"values": map[string]interface{}{
					"buildJob": map[string]interface{}{
						"resources": map[string]interface{}{
							"defaultPreset": "testJobPreser",
						},
					},
					"function": map[string]interface{}{
						"resources": map[string]interface{}{
							"defaultPreset": "testPodPreser",
						},
					},
				},
			},
		}

		flags := NewFlagsBuilder().
			WithNodePort(1234).
			WithDefaultPresetFlags("testJobPreser", "testPodPreser").
			WithOptionalDependencies("testPublisherURL", "testCollectorURL").
			WithRegistryAddresses("testRegistryAddress", "testServerAddress").
			WithRegistryCredentials("testUsername", "testPassword").
			WithRegistryEnableInternal(false).
			WithRegistryHttpSecret("testHttpSecret").
			WithControllerConfiguration(
				"testCPUUtilizationPercentage",
				"testRequeueDuration",
				"testBuildExecutorArgs",
				"testMaxSimultaneousJobs",
				"testHealthzLivenessTimeout",
				"testRequestBodyLimitMb",
				"testTimeoutSec",
			).
			WithExcludedNamespaces([]string{"testNamespace", "testNamespace2"}).Build()

		require.Equal(t, expectedFlags, flags)
	})

	t.Run("build registry flags only", func(t *testing.T) {
		expectedFlags := map[string]interface{}{
			"dockerRegistry": map[string]interface{}{
				"password":        "testPassword",
				"registryAddress": "testRegistryAddress",
				"serverAddress":   "testServerAddress",
				"username":        "testUsername",
			},
		}

		flags := NewFlagsBuilder().
			WithRegistryAddresses("testRegistryAddress", "testServerAddress").
			WithRegistryCredentials("testUsername", "testPassword").
			Build()

		require.Equal(t, expectedFlags, flags)
	})

	t.Run("build not empty controller configuration flags only", func(t *testing.T) {
		expectedFlags := map[string]interface{}{
			"containers": map[string]interface{}{
				"manager": map[string]interface{}{
					"configuration": map[string]interface{}{
						"data": map[string]interface{}{
							"functionBuildMaxSimultaneousJobs": "testMaxSimultaneousJobs",
							"functionRequestBodyLimitMb":       "testRequestBodyLimitMb",
							"healthzLivenessTimeout":           "testHealthzLivenessTimeout",
							"targetCPUUtilizationPercentage":   "testCPUUtilizationPercentage",
						},
					},
				},
			},
		}

		flags := NewFlagsBuilder().
			WithControllerConfiguration(
				"testCPUUtilizationPercentage",
				"",
				"",
				"testMaxSimultaneousJobs",
				"testHealthzLivenessTimeout",
				"testRequestBodyLimitMb",
				"",
			).Build()

		require.Equal(t, expectedFlags, flags)
	})
}
