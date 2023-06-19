package registry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	testRegistrySecret = &corev1.Secret{
		//TODO
	}
)

func TestListExternalRegistrySecrets(t *testing.T) {
	t.Run("returns nil when  external registry secret not found", func(t *testing.T) {
		err := ListExternalRegistrySecrets()
		require.NoError(t, err)
	})

	t.Run("returns error when found secrets", func(t *testing.T) {
		ctx := context.Background()
		namespace := "kyma-test"
		client := fake.NewClientBuilder().
			WithRuntimeObjects(testRegistrySecret).
			Build()

		err := ListExternalRegistrySecrets(ctx, client, namespace)
		require.Error(t, err)
		require.ErrorContains(t, err, "TODO")
	})
}
