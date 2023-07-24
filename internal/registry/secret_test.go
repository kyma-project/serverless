package registry

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	testRegistryEmptySecret  = &corev1.Secret{}
	testRegistryFilledSecret = FixServerlessClusterWideExternalRegistrySecret()
)

func Test_GetExternalRegistrySecret(t *testing.T) {
	t.Run("returns nil when external registry secret not found", func(t *testing.T) {
		ctx := context.Background()
		client := fake.NewClientBuilder().
			Build()
		namespace := "some-namespace"

		secret, err := GetServerlessExternalRegistrySecret(ctx, client, namespace)
		require.NoError(t, err)
		require.Nil(t, secret)
	})

	t.Run("returns secret when found it", func(t *testing.T) {
		ctx := context.Background()
		client := fake.NewClientBuilder().
			WithRuntimeObjects(testRegistryFilledSecret).
			Build()
		namespace := testRegistryFilledSecret.Namespace

		secret, err := GetServerlessExternalRegistrySecret(ctx, client, namespace)
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
			name: "without label remote-registry",
			secretInEnvironment: func() *corev1.Secret {
				secret := testRegistryFilledSecret.DeepCopy()
				delete(secret.Labels, ServerlessExternalRegistryLabelRemoteRegistryKey)
				return secret
			}(),
		},
		{
			name: "without label config",
			secretInEnvironment: func() *corev1.Secret {
				secret := testRegistryFilledSecret.DeepCopy()
				delete(secret.Labels, ServerlessExternalRegistryLabelConfigKey)
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

			secret, err := GetServerlessExternalRegistrySecret(ctx, client, namespace)
			require.NoError(t, err)
			require.Nil(t, secret)
		})
	}
}
