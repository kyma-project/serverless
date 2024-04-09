package registry

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	testRegistryFilledSecret = FixClusterWideExternalRegistrySecret()
)

func Test_GetExternalRegistrySecret(t *testing.T) {
	t.Run("returns nil when external registry secret not found", func(t *testing.T) {
		ctx := context.Background()
		client := fake.NewClientBuilder().
			Build()
		namespace := "some-namespace"

		secret, err := GetExternalClusterWideRegistrySecret(ctx, client, namespace)
		require.NoError(t, err)
		require.Nil(t, secret)
	})

	t.Run("returns secret when found it", func(t *testing.T) {
		ctx := context.Background()
		client := fake.NewClientBuilder().
			WithRuntimeObjects(testRegistryFilledSecret).
			Build()
		namespace := testRegistryFilledSecret.Namespace

		secret, err := GetExternalClusterWideRegistrySecret(ctx, client, namespace)
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
				delete(secret.Labels, ExternalRegistryLabelRemoteRegistryKey)
				return secret
			}(),
		},
		{
			name: "without label config",
			secretInEnvironment: func() *corev1.Secret {
				secret := testRegistryFilledSecret.DeepCopy()
				delete(secret.Labels, ExternalRegistryLabelConfigKey)
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

			secret, err := GetExternalClusterWideRegistrySecret(ctx, client, namespace)
			require.NoError(t, err)
			require.Nil(t, secret)
		})
	}
}

func TestListExternalNamespacedScopeSecrets(t *testing.T) {
	t.Run("return empty list when secrets is not found", func(t *testing.T) {
		ctx := context.Background()
		client := fake.NewClientBuilder().
			Build()

		secrets, err := ListExternalNamespacedScopeSecrets(ctx, client)
		require.NoError(t, err)
		require.Empty(t, secrets)
	})

	t.Run("return non empty list when secret is found", func(t *testing.T) {
		ctx := context.Background()

		client := fake.NewClientBuilder().
			WithRuntimeObjects(&corev1.Secret{
				ObjectMeta: v1.ObjectMeta{
					Name:      ExternalRegistrySecretName,
					Namespace: "default",
					Labels: map[string]string{
						ExternalRegistryLabelRemoteRegistryKey: ExternalRegistryLabelRemoteRegistryVal,
					},
				},
			}).
			Build()

		secrets, err := ListExternalNamespacedScopeSecrets(ctx, client)
		require.NoError(t, err)
		require.Len(t, secrets, 1)
	})

	t.Run("return empty list when secret has wrong name", func(t *testing.T) {
		ctx := context.Background()

		client := fake.NewClientBuilder().
			WithRuntimeObjects(&corev1.Secret{
				ObjectMeta: v1.ObjectMeta{
					Name:      "wrong-name",
					Namespace: "default",
					Labels: map[string]string{
						ExternalRegistryLabelRemoteRegistryKey: ExternalRegistryLabelRemoteRegistryVal,
					},
				},
			}).
			Build()

		secrets, err := ListExternalNamespacedScopeSecrets(ctx, client)
		require.NoError(t, err)
		require.Len(t, secrets, 0)
	})

	t.Run("return empty list when secret has wrong labels", func(t *testing.T) {
		ctx := context.Background()

		client := fake.NewClientBuilder().
			WithRuntimeObjects(&corev1.Secret{
				ObjectMeta: v1.ObjectMeta{
					Name:      ExternalRegistrySecretName,
					Namespace: "default",
					Labels: map[string]string{
						ExternalRegistryLabelRemoteRegistryKey: ExternalRegistryLabelRemoteRegistryVal,
						ExternalRegistryLabelConfigKey:         ExternalRegistryLabelConfigVal,
					},
				},
			}).
			Build()

		secrets, err := ListExternalNamespacedScopeSecrets(ctx, client)
		require.NoError(t, err)
		require.Len(t, secrets, 0)
	})
}
