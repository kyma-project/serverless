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
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const testSecretNamespace = "kyma-system"

func TestManifestCache_Delete(t *testing.T) {
	t.Run("delete secret", func(t *testing.T) {
		key := types.NamespacedName{
			Name:      "test-serverless",
			Namespace: testSecretNamespace,
		}
		ctx := context.TODO()
		client := fake.NewClientBuilder().WithRuntimeObjects(
			fixSecretCache(key, nil),
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
			Name:      "test-serverless",
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
			Name:      "test-serverless",
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
			Name:      "test-serverless",
			Namespace: testSecretNamespace,
		}
		ctx := context.TODO()
		client := fake.NewClientBuilder().WithRuntimeObjects(
			fixSecretCache(key, map[string][]byte{
				"customFlags": []byte("{\"flag1\": \"val1\", \"flag2\": \"val2\"}"),
				"manifest":    []byte("schmetterling"),
			}),
		).Build()

		cache := NewSecretManifestCache(client)

		result, err := cache.Get(ctx, key)
		require.NoError(t, err)

		expectedResult := &ServerlessSpecManifest{
			customFlags: map[string]interface{}{
				"flag1": "val1",
				"flag2": "val2",
			},
			manifest: "schmetterling",
		}
		require.Equal(t, expectedResult, result)
	})

	t.Run("client error", func(t *testing.T) {
		scheme := runtime.NewScheme()
		// apiextensionscheme does not contains v1.Secret scheme
		require.NoError(t, apiextensionsscheme.AddToScheme(scheme))

		key := types.NamespacedName{
			Name:      "test-serverless",
			Namespace: testSecretNamespace,
		}
		ctx := context.TODO()
		client := fake.NewClientBuilder().WithScheme(scheme).Build()

		cache := NewSecretManifestCache(client)

		result, err := cache.Get(ctx, key)
		require.Error(t, err)
		require.Nil(t, result)
	})

	t.Run("secret not found", func(t *testing.T) {
		key := types.NamespacedName{
			Name:      "test-serverless",
			Namespace: testSecretNamespace,
		}
		ctx := context.TODO()
		client := fake.NewClientBuilder().Build()

		cache := NewSecretManifestCache(client)

		result, err := cache.Get(ctx, key)
		require.NoError(t, err)
		require.Nil(t, result)
	})

	t.Run("conversion error", func(t *testing.T) {
		key := types.NamespacedName{
			Name:      "test-serverless",
			Namespace: testSecretNamespace,
		}
		ctx := context.TODO()
		client := fake.NewClientBuilder().WithRuntimeObjects(
			fixSecretCache(key, map[string][]byte{
				"customFlags": []byte("{UNEXPECTED}"),
				"manifest":    []byte("schmetterling"),
			}),
		).Build()

		cache := NewSecretManifestCache(client)

		result, err := cache.Get(ctx, key)
		require.Error(t, err)
		require.Nil(t, result)
	})
}

func TestManifestCache_Set(t *testing.T) {
	t.Run("create secret", func(t *testing.T) {
		key := types.NamespacedName{
			Name:      "test-serverless",
			Namespace: testSecretNamespace,
		}
		ctx := context.TODO()
		client := fake.NewClientBuilder().Build()
		manifest := "schmetterling"
		flags := map[string]interface{}{
			"flag1": "val1",
			"flag2": "val2",
		}

		cache := NewSecretManifestCache(client)

		err := cache.Set(ctx, key, flags, manifest)
		require.NoError(t, err)

		var secret corev1.Secret
		require.NoError(t, client.Get(ctx, key, &secret))
		require.Equal(t, []byte("schmetterling"), secret.Data["manifest"])
		require.Equal(t, []byte("{\"flag1\":\"val1\",\"flag2\":\"val2\"}"), secret.Data["customFlags"])
	})

	t.Run("update secret", func(t *testing.T) {
		key := types.NamespacedName{
			Name:      "test-serverless",
			Namespace: testSecretNamespace,
		}
		ctx := context.TODO()
		client := fake.NewClientBuilder().WithRuntimeObjects(
			fixSecretCache(key, nil),
		).Build()
		manifest := "schmetterling"
		flags := map[string]interface{}{
			"flag1": "val1",
			"flag2": "val2",
		}

		cache := NewSecretManifestCache(client)

		err := cache.Set(ctx, key, flags, manifest)
		require.NoError(t, err)

		var secret corev1.Secret
		require.NoError(t, client.Get(ctx, key, &secret))
		require.Equal(t, []byte("schmetterling"), secret.Data["manifest"])
		require.Equal(t, []byte("{\"flag1\":\"val1\",\"flag2\":\"val2\"}"), secret.Data["customFlags"])
	})

	t.Run("marshal error", func(t *testing.T) {
		key := types.NamespacedName{
			Name:      "test-serverless",
			Namespace: testSecretNamespace,
		}
		ctx := context.TODO()
		client := fake.NewClientBuilder().Build()
		wrongFlags := map[string]interface{}{
			"flag1": func() {},
		}

		cache := NewSecretManifestCache(client)

		err := cache.Set(ctx, key, wrongFlags, "")
		require.Error(t, err)
	})
}

func fixSecretCache(key types.NamespacedName, data map[string][]byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Data: data,
	}
}
