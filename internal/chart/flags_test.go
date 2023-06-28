package chart

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	testRegistrySecret = corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Data: map[string][]byte{
			"username":        []byte("test-username"),
			"password":        []byte("test-password"),
			"registryAddress": []byte("test-registryAddress"),
			"serverAddress":   []byte("test-serverAddress"),
		},
	}
)

// TODO: add tests

func TestAppendContainersFlags(t *testing.T) {
	t.Run("append flags", func(t *testing.T) {
		publisherURL := "test-proxy-url"
		collectorURL := "test-trace-url"

		flags := AppendContainersFlags(map[string]interface{}{}, publisherURL, collectorURL)

		require.Equal(t, map[string]interface{}{
			"containers": map[string]interface{}{
				"manager": map[string]interface{}{
					"envs": map[string]interface{}{
						"functionTraceCollectorEndpoint": map[string]interface{}{
							"value": collectorURL,
						},
						"functionPublisherProxyAddress": map[string]interface{}{
							"value": publisherURL,
						},
					},
				},
			},
		}, flags)
	})
}
