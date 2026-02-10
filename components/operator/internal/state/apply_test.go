package state

import (
	"context"
	"testing"

	"github.com/kyma-project/manager-toolkit/installation/chart"
	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/components/operator/internal/flags"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	envName             = "IMAGE_FUNCTION_CONTROLLER"
	runtimeImageEnvName = "IMAGE_FUNCTION_RUNTIME_NODEJS24"
)

func Test_buildSFnApplyResources(t *testing.T) {
	t.Run("switch state and add condition when condition is missing", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{},
			chartConfig: &chart.Config{
				Cache: fixEmptyManifestCache(),
				CacheKey: types.NamespacedName{
					Name:      testInstalledServerless.GetName(),
					Namespace: testInstalledServerless.GetNamespace(),
				},
				Release: chart.Release{
					Name:      testInstalledServerless.GetName(),
					Namespace: testInstalledServerless.GetNamespace(),
				},
			},
			flagsBuilder: flags.NewBuilder(),
		}

		r := &reconciler{
			k8s: k8s{
				client: fake.NewClientBuilder().Build(),
			},
		}

		next, result, err := sFnApplyResources(context.Background(), r, s)
		require.Nil(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnVerifyResources, next)

		expectedFlags := map[string]interface{}{
			"global": map[string]interface{}{
				"commonLabels": map[string]interface{}{
					"app.kubernetes.io/managed-by": "serverless-operator",
				},
			},
		}

		flags, err := s.flagsBuilder.Build()
		require.NoError(t, err)

		require.Equal(t, expectedFlags, flags)

		status := s.instance.Status
		require.Equal(t, v1alpha1.StateProcessing, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeInstalled,
			metav1.ConditionUnknown,
			v1alpha1.ConditionReasonInstallation,
			"Installing for configuration",
		)
	})

	t.Run("apply resources", func(t *testing.T) {
		s := &systemState{
			instance: *testInstalledServerless.DeepCopy(),
			chartConfig: &chart.Config{
				Cache: fixEmptyManifestCache(),
				CacheKey: types.NamespacedName{
					Name:      testInstalledServerless.GetName(),
					Namespace: testInstalledServerless.GetNamespace(),
				},
				Release: chart.Release{
					Name:      testInstalledServerless.GetName(),
					Namespace: testInstalledServerless.GetNamespace(),
				},
			},
			flagsBuilder: flags.NewBuilder(),
		}
		r := &reconciler{}

		// run installation process and return verificating state
		next, result, err := sFnApplyResources(context.Background(), r, s)
		require.Nil(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnVerifyResources, next)
	})

	t.Run("install chart error", func(t *testing.T) {
		s := &systemState{
			instance: *testInstalledServerless.DeepCopy(),
			chartConfig: &chart.Config{
				Cache: fixManifestCache("\t"),
				CacheKey: types.NamespacedName{
					Name:      testInstalledServerless.GetName(),
					Namespace: testInstalledServerless.GetNamespace(),
				},
			},
			flagsBuilder: flags.NewBuilder(),
		}
		r := &reconciler{
			log: zap.NewNop().Sugar(),
		}

		// handle error and return update condition state
		next, result, err := sFnApplyResources(context.Background(), r, s)
		require.EqualError(t, err, "could not parse chart manifest: yaml: found character that cannot start any token")
		require.Nil(t, result)
		require.Nil(t, next)

		status := s.instance.Status
		require.Equal(t, v1alpha1.StateError, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeInstalled,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonInstallationErr,
			"could not parse chart manifest: yaml: found character that cannot start any token",
		)
	})
}

func TestUpdateImageIfOverride(t *testing.T) {
	type caseDef struct {
		name          string
		envKey        string
		envs          map[string]string
		bindUpdater   func(*flags.Builder) flags.ImageReplace
		expectedKey   string
		expectedValue string
	}

	cases := []caseDef{
		{
			name:          "Override image",
			envKey:        envName,
			envs:          map[string]string{envName: "newImage"},
			bindUpdater:   func(b *flags.Builder) flags.ImageReplace { return b.WithImageFunctionBuildfulController },
			expectedKey:   "function_buildful_controller",
			expectedValue: "newImage",
		},
		{
			name:        "Don't override image when empty env",
			envKey:      envName,
			envs:        map[string]string{},
			bindUpdater: func(b *flags.Builder) flags.ImageReplace { return b.WithImageFunctionBuildfulController },
		},
		{
			name:          "Choose FIPS variant images when Fips mode enabled",
			envKey:        runtimeImageEnvName,
			envs:          map[string]string{runtimeImageEnvName: "non-fips-image", runtimeImageEnvName + fipsVariantImageEnvKeySuffix: "fips-image", kymaFipsModeEnv: "true"},
			bindUpdater:   func(b *flags.Builder) flags.ImageReplace { return b.WithImageFunctionRuntimeNodejs24 },
			expectedKey:   "function_runtime_nodejs24",
			expectedValue: "fips-image",
		},
		{
			name:          "Fallback to non-FIPS when FIPS variant not set",
			envKey:        runtimeImageEnvName,
			envs:          map[string]string{runtimeImageEnvName: "non-fips-image", kymaFipsModeEnv: "true"},
			bindUpdater:   func(b *flags.Builder) flags.ImageReplace { return b.WithImageFunctionRuntimeNodejs24 },
			expectedKey:   "function_runtime_nodejs24",
			expectedValue: "non-fips-image",
		},
		{
			name:          "Use non-FIPS when Fips flag not set",
			envKey:        runtimeImageEnvName,
			envs:          map[string]string{runtimeImageEnvName: "non-fips-image", runtimeImageEnvName + fipsVariantImageEnvKeySuffix: "fips-image"},
			bindUpdater:   func(b *flags.Builder) flags.ImageReplace { return b.WithImageFunctionRuntimeNodejs24 },
			expectedKey:   "function_runtime_nodejs24",
			expectedValue: "non-fips-image",
		},
		{
			name:          "Use non-FIPS when Fips flag non-true",
			envKey:        runtimeImageEnvName,
			envs:          map[string]string{runtimeImageEnvName: "non-fips-image", runtimeImageEnvName + fipsVariantImageEnvKeySuffix: "fips-image", kymaFipsModeEnv: "makapaka"},
			bindUpdater:   func(b *flags.Builder) flags.ImageReplace { return b.WithImageFunctionRuntimeNodejs24 },
			expectedKey:   "function_runtime_nodejs24",
			expectedValue: "non-fips-image",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			// set provided envs using t.Setenv so they are cleaned up automatically
			for k, v := range tc.envs {
				t.Setenv(k, v)
			}

			fb := flags.NewBuilder()
			updater := tc.bindUpdater(fb)

			updateImageIfOverride(tc.envKey, updater)

			flagsMap, err := fb.Build()
			require.NoError(t, err)

			// build expected flags
			var expected map[string]interface{}
			if tc.expectedKey != "" {
				expected = map[string]interface{}{
					"global": map[string]interface{}{
						"images": map[string]interface{}{
							tc.expectedKey: tc.expectedValue,
						},
					},
				}
			} else {
				expected = map[string]interface{}{}
			}

			require.Equal(t, expected, flagsMap)
		})
	}
}
