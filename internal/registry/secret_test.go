package registry

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	testRegistryEmptySecret  = &corev1.Secret{}
	testRegistryFilledSecret = &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "serverless-registry-config",
			Namespace: "kyma-test",
			Labels: map[string]string{
				"serverless.kyma-project.io/remote-registry": "config",
				"serverless.kyma-project.io/config":          "credentials",
			},
		},
		Type: corev1.SecretTypeDockerConfigJson,
	}
)

func TestListExternalRegistrySecrets(t *testing.T) {
	t.Run("returns nil when  external registry secret not found", func(t *testing.T) {
		ctx := context.Background()
		client := fake.NewClientBuilder().
			WithRuntimeObjects(testRegistryEmptySecret).
			Build()

		err := DetectExternalRegistrySecrets(ctx, client)
		require.NoError(t, err)
	})

	t.Run("returns error when found secrets", func(t *testing.T) {
		ctx := context.Background()
		client := fake.NewClientBuilder().
			WithRuntimeObjects(testRegistryFilledSecret).
			Build()

		err := DetectExternalRegistrySecrets(ctx, client)
		require.Error(t, err)
		require.ErrorContains(t, err, "test-secret")
	})
}

func Test_GetExternalRegistrySecret(t *testing.T) {
	t.Run("returns nil when external registry secret not found", func(t *testing.T) {
		ctx := context.Background()
		client := fake.NewClientBuilder().
			Build()
		namespace := "some-namespace"

		secret, err := GetExternalRegistrySecret(ctx, client, namespace)
		require.NoError(t, err)
		require.Nil(t, secret)
	})

	t.Run("returns secret when found it", func(t *testing.T) {
		ctx := context.Background()
		client := fake.NewClientBuilder().
			WithRuntimeObjects(testRegistryFilledSecret).
			Build()
		namespace := testRegistryFilledSecret.Namespace

		secret, err := GetExternalRegistrySecret(ctx, client, namespace)
		require.NoError(t, err)
		require.NotNil(t, secret)
		require.Equal(t, testRegistryFilledSecret, secret)
	})

	noProperSecretTests := []struct {
		name                string
		secretInEnvironment *corev1.Secret
	}{
		{
			name: "bad name",
			secretInEnvironment: func() *corev1.Secret {
				secret := testRegistryFilledSecret.DeepCopy()
				secret.Name = "bad-name"
				return secret
			}(),
		},
		{
			name: "bad type",
			secretInEnvironment: func() *corev1.Secret {
				secret := testRegistryFilledSecret.DeepCopy()
				secret.Type = corev1.SecretTypeBasicAuth
				return secret
			}(),
		},
		{
			name: "without label remote-registry",
			secretInEnvironment: func() *corev1.Secret {
				secret := testRegistryFilledSecret.DeepCopy()
				delete(secret.Labels, "serverless.kyma-project.io/remote-registry")
				return secret
			}(),
		},
		{
			name: "without label config",
			secretInEnvironment: func() *corev1.Secret {
				secret := testRegistryFilledSecret.DeepCopy()
				delete(secret.Labels, "serverless.kyma-project.io/config")
				return secret
			}(),
		},
	}
	for _, tt := range noProperSecretTests {
		t.Run(fmt.Sprintf("returns nil when no secret has proper params - %s", tt.name), func(t *testing.T) {
			ctx := context.Background()
			client := fake.NewClientBuilder().
				WithObjects(tt.secretInEnvironment).
				Build()
			namespace := testRegistryFilledSecret.Namespace

			secret, err := GetExternalRegistrySecret(ctx, client, namespace)
			require.NoError(t, err)
			require.Nil(t, secret)
		})
	}
}
