package state

import (
	"context"
	"errors"
	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func Test_sFnRegistryConfiguration(t *testing.T) {
	t.Run("internal registry and update", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{
				Spec: v1alpha1.ServerlessSpec{
					DockerRegistry: &v1alpha1.DockerRegistry{
						EnableInternal: pointer.Bool(true),
					},
				},
			},
			snapshot: v1alpha1.ServerlessStatus{
				DockerRegistry: "",
			},
			chartConfig: &chart.Config{
				Release: chart.Release{
					Flags: chart.EmptyFlags(),
				},
			},
		}

		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().Build()},
			log: zap.NewNop().Sugar(),
		}

		expectedFlags := map[string]interface{}{
			"dockerRegistry": map[string]interface{}{
				"enableInternal": true,
			},
			"global": map[string]interface{}{
				"registryNodePort": int64(32_137),
			},
		}
		expectedNext := sFnUpdateStatusAndRequeue

		next, result, err := sFnRegistryConfiguration(context.Background(), r, s)
		require.Nil(t, result)
		require.NoError(t, err)
		requireEqualFunc(t, expectedNext, next)

		require.EqualValues(t, expectedFlags, s.chartConfig.Release.Flags)
		require.Equal(t, "internal", s.instance.Status.DockerRegistry)
	})
	t.Run("external registry and go to next state", func(t *testing.T) {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-secret",
				Namespace: "kyma-test",
			},
			Data: map[string][]byte{
				"username":        []byte("username"),
				"password":        []byte("password"),
				"registryAddress": []byte("registryAddress"),
				"serverAddress":   []byte("serverAddress"),
			},
		}

		s := &systemState{
			instance: v1alpha1.Serverless{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kyma-test",
				},
				Spec: v1alpha1.ServerlessSpec{
					DockerRegistry: &v1alpha1.DockerRegistry{
						EnableInternal: pointer.Bool(false),
						SecretName:     pointer.String("test-secret"),
					},
				},
			},
			snapshot: v1alpha1.ServerlessStatus{
				DockerRegistry: string(secret.Data["serverAddress"]),
			},
			chartConfig: &chart.Config{
				Release: chart.Release{
					Flags: chart.EmptyFlags(),
				},
			},
		}
		r := &reconciler{
			k8s: k8s{
				client: fake.NewClientBuilder().
					WithRuntimeObjects(secret).
					Build(),
			},
		}
		expectedFlags := map[string]interface{}{
			"dockerRegistry": map[string]interface{}{
				"enableInternal":  false,
				"username":        string(secret.Data["username"]),
				"password":        string(secret.Data["password"]),
				"registryAddress": string(secret.Data["registryAddress"]),
				"serverAddress":   string(secret.Data["serverAddress"]),
			},
		}
		expectedNext := sFnOptionalDependencies

		next, result, err := sFnRegistryConfiguration(context.Background(), r, s)
		require.Nil(t, result)
		require.NoError(t, err)
		requireEqualFunc(t, expectedNext, next)

		require.Equal(t, expectedFlags, s.chartConfig.Release.Flags)
		require.Equal(t, string(secret.Data["serverAddress"]), s.instance.Status.DockerRegistry)
	})
	t.Run("k3d registry and update", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{
				Spec: v1alpha1.ServerlessSpec{
					DockerRegistry: &v1alpha1.DockerRegistry{
						EnableInternal: pointer.Bool(false),
					},
				},
			},
			snapshot: v1alpha1.ServerlessStatus{
				DockerRegistry: "",
			},
			chartConfig: &chart.Config{
				Release: chart.Release{
					Flags: chart.EmptyFlags(),
				},
			},
		}
		expectedFlags := map[string]interface{}{
			"dockerRegistry": map[string]interface{}{
				"enableInternal":  false,
				"registryAddress": v1alpha1.DefaultRegistryAddress,
				"serverAddress":   v1alpha1.DefaultRegistryAddress,
			},
		}
		expectedNext := sFnUpdateStatusAndRequeue

		next, result, err := sFnRegistryConfiguration(context.Background(), nil, s)
		require.Nil(t, result)
		require.NoError(t, err)
		requireEqualFunc(t, expectedNext, next)

		require.Equal(t, expectedFlags, s.chartConfig.Release.Flags)
		require.Equal(t, v1alpha1.DefaultRegistryAddress, s.instance.Status.DockerRegistry)
	})
	t.Run("external registry secret not found error", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{
				Spec: v1alpha1.ServerlessSpec{
					DockerRegistry: &v1alpha1.DockerRegistry{
						EnableInternal: pointer.Bool(false),
						SecretName:     pointer.String("test-secret-not-found"),
					},
				},
			},
		}
		r := &reconciler{
			k8s: k8s{
				client: fake.NewClientBuilder().Build(),
			},
		}
		expectedNext := sFnUpdateStatusWithError(errors.New("test error"))

		next, result, err := sFnRegistryConfiguration(context.Background(), r, s)
		require.Nil(t, result)
		require.NoError(t, err)
		requireEqualFunc(t, expectedNext, next)

		status := s.instance.Status
		require.Equal(t, v1alpha1.StateError, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeConfigured,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonConfigurationErr,
			"secrets \"test-secret-not-found\" not found",
		)
	})
}
