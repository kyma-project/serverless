package registry

import (
	"context"
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
			Kind:       "secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "kyma-test",
			Labels: map[string]string{
				"serverless.kyma-project.io/remote-registry": "config",
			},
		},
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
