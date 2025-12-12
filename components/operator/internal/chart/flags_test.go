package chart

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_flagsBuilder_Build(t *testing.T) {
	t.Run("build empty flags", func(t *testing.T) {
		flags, err := NewFlagsBuilder().Build()
		require.NoError(t, err)
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
							"functionRequeueDuration":          "testRequeueDuration",
							"functionTraceCollectorEndpoint":   "testCollectorURL",
							"healthzLivenessTimeout":           "testHealthzLivenessTimeout",
							"targetCPUUtilizationPercentage":   "testCPUUtilizationPercentage",
							"resourcesConfiguration": map[string]interface{}{
								"function": map[string]interface{}{
									"resources": map[string]interface{}{
										"defaultPreset": "testPodPreset",
									},
								},
								"buildJob": map[string]interface{}{
									"resources": map[string]interface{}{
										"defaultPreset": "testJobPreset",
									},
								},
							},
						},
					},
					"logConfiguration": map[string]interface{}{
						"data": map[string]interface{}{
							"logLevel":  "testLogLevel",
							"logFormat": "testLogFormat",
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
				"username":        "testUsername",
			},
			"global": map[string]interface{}{
				"commonLabels": map[string]interface{}{
					"app.kubernetes.io/managed-by": "test-runner",
				},
				"registryNodePort": int64(1234),
			},
		}

		flags, err := NewFlagsBuilder().
			WithNodePort(1234).
			WithDefaultPresetFlags("testJobPreset", "testPodPreset").
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
			).
			WithLogFormat("testLogFormat").
			WithLogLevel("testLogLevel").
			WithManagedByLabel("test-runner").Build()

		require.NoError(t, err)
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

		flags, err := NewFlagsBuilder().
			WithRegistryAddresses("testRegistryAddress", "testServerAddress").
			WithRegistryCredentials("testUsername", "testPassword").
			Build()

		require.NoError(t, err)
		require.Equal(t, expectedFlags, flags)
	})

	t.Run("build not empty controller configuration flags only", func(t *testing.T) {
		expectedFlags := map[string]interface{}{
			"containers": map[string]interface{}{
				"manager": map[string]interface{}{
					"configuration": map[string]interface{}{
						"data": map[string]interface{}{
							"functionBuildMaxSimultaneousJobs": "testMaxSimultaneousJobs",
							"healthzLivenessTimeout":           "testHealthzLivenessTimeout",
							"targetCPUUtilizationPercentage":   "testCPUUtilizationPercentage",
						},
					},
				},
			},
		}

		flags, err := NewFlagsBuilder().
			WithControllerConfiguration(
				"testCPUUtilizationPercentage",
				"",
				"",
				"testMaxSimultaneousJobs",
				"testHealthzLivenessTimeout",
			).Build()

		require.NoError(t, err)
		require.Equal(t, expectedFlags, flags)
	})

	t.Run("build images flags only", func(t *testing.T) {
		expectedFlags := map[string]interface{}{
			"global": map[string]interface{}{
				"images": map[string]interface{}{
					"function_buildful_controller": "testFunctionBuildfulController",
					"function_controller":          "testFunctionController",
					"function_build_init":          "testFunctionBuildInit",
					"function_init":                "testFunctionInit",
					"registry_init":                "testRegistryInit",
					"function_runtime_nodejs20":    "testFunctionRuntimeNodejs20",
					"function_runtime_nodejs22":    "testFunctionRuntimeNodejs22",
					"function_runtime_python312":   "testFunctionRuntimePython312",
					"kaniko_executor":              "testKanikoExecutor",
					"registry":                     "testRegistry",
				},
			},
		}

		flags, err := NewFlagsBuilder().
			WithImageFunctionBuildfulController("testFunctionBuildfulController").
			WithImageFunctionController("testFunctionController").
			WithImageFunctionBuildInit("testFunctionBuildInit").
			WithImageFunctionInit("testFunctionInit").
			WithImageRegistryInit("testRegistryInit").
			WithImageFunctionRuntimeNodejs20("testFunctionRuntimeNodejs20").
			WithImageFunctionRuntimeNodejs22("testFunctionRuntimeNodejs22").
			WithImageFunctionRuntimePython312("testFunctionRuntimePython312").
			WithImageKanikoExecutor("testKanikoExecutor").
			WithImageRegistry("testRegistry").
			Build()

		require.NoError(t, err)
		require.Equal(t, expectedFlags, flags)
	})
}
