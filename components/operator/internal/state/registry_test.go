package state

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/components/operator/internal/chart"
	"github.com/kyma-project/serverless/components/operator/internal/warning"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_sFnRegistryConfiguration(t *testing.T) {
	t.Run("no internal registry status", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{
				Spec: v1alpha1.ServerlessSpec{
					DockerRegistry: &v1alpha1.DockerRegistry{
						EnableInternal: ptr.To[bool](true),
					},
				},
			},
			statusSnapshot: v1alpha1.ServerlessStatus{
				DockerRegistry: "",
			},
			flagsBuilder: chart.NewFlagsBuilder(),
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

		next, result, err := sFnRegistryConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnOptionalDependencies, next)

		flags, err := s.flagsBuilder.Build()
		require.NoError(t, err)

		require.EqualValues(t, expectedFlags, flags)
		require.Equal(t, "", s.instance.Status.DockerRegistry)
		require.Equal(t, v1alpha1.StateProcessing, s.instance.Status.State)
	})

	t.Run("internal registry and update", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						buildlessModeAnnotation: buildlessModeDisabled,
					},
				},
				Spec: v1alpha1.ServerlessSpec{
					DockerRegistry: &v1alpha1.DockerRegistry{
						EnableInternal: ptr.To[bool](true),
					},
				},
			},
			statusSnapshot: v1alpha1.ServerlessStatus{
				DockerRegistry: "",
			},
			flagsBuilder: chart.NewFlagsBuilder(),
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

		next, result, err := sFnRegistryConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnOptionalDependencies, next)

		flags, err := s.flagsBuilder.Build()
		require.NoError(t, err)

		require.EqualValues(t, expectedFlags, flags)
		require.Equal(t, "internal", s.instance.Status.DockerRegistry)
		require.Equal(t, v1alpha1.StateProcessing, s.instance.Status.State)
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
						EnableInternal: ptr.To[bool](false),
						SecretName:     ptr.To[string]("test-secret"),
					},
				},
			},
			statusSnapshot: v1alpha1.ServerlessStatus{
				DockerRegistry: string(secret.Data["serverAddress"]),
			},
			flagsBuilder: chart.NewFlagsBuilder(),
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

		next, result, err := sFnRegistryConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnOptionalDependencies, next)

		flags, err := s.flagsBuilder.Build()
		require.NoError(t, err)

		require.Equal(t, expectedFlags, flags)
		require.Equal(t, string(secret.Data["serverAddress"]), s.instance.Status.DockerRegistry)
		require.Equal(t, v1alpha1.StateProcessing, s.instance.Status.State)
	})

	t.Run("k3d registry and update", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "some-namespace",
					Annotations: map[string]string{
						buildlessModeAnnotation: buildlessModeDisabled,
					},
				},
				Spec: v1alpha1.ServerlessSpec{
					DockerRegistry: &v1alpha1.DockerRegistry{
						EnableInternal: ptr.To[bool](false),
					},
				},
			},
			statusSnapshot: v1alpha1.ServerlessStatus{
				DockerRegistry: "",
			},
			flagsBuilder: chart.NewFlagsBuilder(),
		}
		r := &reconciler{
			k8s: k8s{
				client: fake.NewClientBuilder().Build(),
			},
		}
		expectedFlags := map[string]interface{}{
			"dockerRegistry": map[string]interface{}{
				"enableInternal":  false,
				"registryAddress": v1alpha1.DefaultRegistryAddress,
				"serverAddress":   v1alpha1.DefaultRegistryAddress,
			},
		}

		next, result, err := sFnRegistryConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnOptionalDependencies, next)

		flags, err := s.flagsBuilder.Build()
		require.NoError(t, err)

		require.Equal(t, expectedFlags, flags)
		require.Equal(t, v1alpha1.DefaultRegistryAddress, s.instance.Status.DockerRegistry)
	})

	t.Run("external registry secret not found error", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "some-namespace",
				},
				Spec: v1alpha1.ServerlessSpec{
					DockerRegistry: &v1alpha1.DockerRegistry{
						EnableInternal: ptr.To[bool](false),
						SecretName:     ptr.To[string]("test-secret-not-found"),
					},
				},
			},
		}
		r := &reconciler{
			k8s: k8s{
				client: fake.NewClientBuilder().Build(),
			},
		}

		next, result, err := sFnRegistryConfiguration(context.Background(), r, s)
		require.EqualError(t, err, "secrets \"test-secret-not-found\" not found")
		require.Nil(t, result)
		require.Nil(t, next)

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

func Test_addRegistryConfigurationWarnings(t *testing.T) {

	t.Run("enable internal is true and secret name exists", func(t *testing.T) {
		s := &systemState{
			warningBuilder: warning.NewBuilder(),
			instance: v1alpha1.Serverless{
				Spec: v1alpha1.ServerlessSpec{
					DockerRegistry: &v1alpha1.DockerRegistry{
						EnableInternal: ptr.To[bool](true),
						SecretName:     ptr.To[string]("test-secret"),
					},
				},
			},
		}
		addRegistryConfigurationWarnings(s)
		require.Equal(t, fmt.Sprintf("Warning: %s", internalEnabledAndSecretNameUsedMessage), s.warningBuilder.Build())
	})

	t.Run("do not build warning", func(t *testing.T) {
		s := &systemState{
			warningBuilder: warning.NewBuilder(),
			instance: v1alpha1.Serverless{
				Spec: v1alpha1.ServerlessSpec{
					DockerRegistry: &v1alpha1.DockerRegistry{
						EnableInternal: ptr.To[bool](false),
						SecretName:     ptr.To[string]("serverless-registry-config"),
					},
				},
			},
		}
		addRegistryConfigurationWarnings(s)
		require.Equal(t, "", s.warningBuilder.Build())
	})
}
