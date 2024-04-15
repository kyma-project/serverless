package chart

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apiextensionsscheme "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const testSecretNamespace = "kyma-system"

func TestManifestCache_Delete(t *testing.T) {
	t.Run("delete secret", func(t *testing.T) {
		key := types.NamespacedName{
			Name:      "test-name",
			Namespace: testSecretNamespace,
		}
		ctx := context.TODO()
		client := fake.NewClientBuilder().WithRuntimeObjects(
			fixSecretCache(t, key, emptySpecManifest),
		).Build()

		cache := NewSecretManifestCache(client)

		err := cache.Delete(ctx, key)
		require.NoError(t, err)

		var secret corev1.Secret
		err = client.Get(ctx, key, &secret)
		require.True(t, errors.IsNotFound(err), fmt.Sprintf("got error: %v", err))
	})

	t.Run("delete error", func(t *testing.T) {
		scheme := runtime.NewScheme()
		// apiextensionscheme does not contains v1.Secret scheme
		require.NoError(t, apiextensionsscheme.AddToScheme(scheme))

		key := types.NamespacedName{
			Name:      "test-name",
			Namespace: testSecretNamespace,
		}
		ctx := context.TODO()
		client := fake.NewClientBuilder().WithScheme(scheme).Build()

		cache := NewSecretManifestCache(client)

		err := cache.Delete(ctx, key)
		require.Error(t, err)
	})

	t.Run("do nothing when cache is not found", func(t *testing.T) {
		key := types.NamespacedName{
			Name:      "test-name",
			Namespace: testSecretNamespace,
		}
		ctx := context.TODO()
		client := fake.NewClientBuilder().Build()

		cache := NewSecretManifestCache(client)

		err := cache.Delete(ctx, key)
		require.NoError(t, err)
	})
}

func TestManifestCache_Get(t *testing.T) {
	t.Run("get secret value", func(t *testing.T) {
		key := types.NamespacedName{
			Name:      "test-name",
			Namespace: testSecretNamespace,
		}
		ctx := context.TODO()
		client := fake.NewClientBuilder().WithRuntimeObjects(
			fixSecretCache(t, key, DockerRegistrySpecManifest{
				CustomFlags: map[string]interface{}{
					"flag1": "val1",
					"flag2": "val2",
				},
				Manifest: "schmetterling",
			}),
		).Build()

		cache := NewSecretManifestCache(client)

		result, err := cache.Get(ctx, key)
		require.NoError(t, err)

		expectedResult := DockerRegistrySpecManifest{
			CustomFlags: map[string]interface{}{
				"flag1": "val1",
				"flag2": "val2",
			},
			Manifest: "schmetterling",
		}
		require.Equal(t, expectedResult, result)
	})

	t.Run("client error", func(t *testing.T) {
		scheme := runtime.NewScheme()
		// apiextensionscheme does not contains v1.Secret scheme
		require.NoError(t, apiextensionsscheme.AddToScheme(scheme))

		key := types.NamespacedName{
			Name:      "test-name",
			Namespace: testSecretNamespace,
		}
		ctx := context.TODO()
		client := fake.NewClientBuilder().WithScheme(scheme).Build()

		cache := NewSecretManifestCache(client)

		result, err := cache.Get(ctx, key)
		require.Error(t, err)
		require.Equal(t, emptySpecManifest, result)
	})

	t.Run("secret not found", func(t *testing.T) {
		key := types.NamespacedName{
			Name:      "test-name",
			Namespace: testSecretNamespace,
		}
		ctx := context.TODO()
		client := fake.NewClientBuilder().Build()

		cache := NewSecretManifestCache(client)

		result, err := cache.Get(ctx, key)
		require.NoError(t, err)
		require.Equal(t, emptySpecManifest, result)
	})

	t.Run("conversion error", func(t *testing.T) {
		key := types.NamespacedName{
			Name:      "test-name",
			Namespace: testSecretNamespace,
		}
		ctx := context.TODO()
		client := fake.NewClientBuilder().WithRuntimeObjects(
			&corev1.Secret{
				ObjectMeta: v1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Data: map[string][]byte{
					"spec": []byte("{UNEXPECTED}"),
				},
			}).Build()

		cache := NewSecretManifestCache(client)

		result, err := cache.Get(ctx, key)
		require.Error(t, err)
		require.Equal(t, emptySpecManifest, result)
	})
}

func TestManifestCache_Set(t *testing.T) {
	t.Run("create secret", func(t *testing.T) {
		key := types.NamespacedName{
			Name:      "test-name",
			Namespace: testSecretNamespace,
		}
		ctx := context.TODO()
		client := fake.NewClientBuilder().Build()

		cache := NewSecretManifestCache(client)
		expectedSpec := DockerRegistrySpecManifest{
			Manifest: "schmetterling",
			CustomFlags: map[string]interface{}{
				"flag1": "val1",
				"flag2": "val2",
			},
		}

		err := cache.Set(ctx, key, expectedSpec)
		require.NoError(t, err)

		var secret corev1.Secret
		require.NoError(t, client.Get(ctx, key, &secret))

		actualSpec := DockerRegistrySpecManifest{}
		err = json.Unmarshal(secret.Data["spec"], &actualSpec)
		require.NoError(t, err)

		require.Equal(t, expectedSpec, actualSpec)
	})

	t.Run("update secret", func(t *testing.T) {
		key := types.NamespacedName{
			Name:      "test-name",
			Namespace: testSecretNamespace,
		}
		ctx := context.TODO()
		client := fake.NewClientBuilder().WithRuntimeObjects(
			fixSecretCache(t, key, emptySpecManifest),
		).Build()

		cache := NewSecretManifestCache(client)
		expectedSpec := DockerRegistrySpecManifest{
			Manifest: "schmetterling",
			CustomFlags: map[string]interface{}{
				"flag1": "val1",
				"flag2": "val2",
			},
		}
		err := cache.Set(ctx, key, expectedSpec)
		require.NoError(t, err)

		var secret corev1.Secret
		require.NoError(t, client.Get(ctx, key, &secret))

		actualSpec := DockerRegistrySpecManifest{}
		err = json.Unmarshal(secret.Data["spec"], &actualSpec)
		require.NoError(t, err)

		require.Equal(t, expectedSpec, actualSpec)
	})

	t.Run("marshal error", func(t *testing.T) {
		key := types.NamespacedName{
			Name:      "test-name",
			Namespace: testSecretNamespace,
		}
		ctx := context.TODO()
		client := fake.NewClientBuilder().Build()
		wrongFlags := map[string]interface{}{
			"flag1": func() {},
		}

		cache := NewSecretManifestCache(client)

		err := cache.Set(ctx, key, DockerRegistrySpecManifest{
			Manifest:    "",
			CustomFlags: wrongFlags,
		})
		require.Error(t, err)
	})
}

func fixSecretCache(t *testing.T, key types.NamespacedName, spec DockerRegistrySpecManifest) *corev1.Secret {
	byteSpec, err := json.Marshal(&spec)
	require.NoError(t, err)

	return &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Data: map[string][]byte{
			"spec": byteSpec,
		},
	}
}
