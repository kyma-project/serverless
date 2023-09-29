package registry

import (
	"k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func FixServerlessClusterWideExternalRegistrySecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: v12.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: v12.ObjectMeta{
			Name:      ServerlessExternalRegistrySecretName,
			Namespace: "kyma-test",
			Labels: map[string]string{
				ServerlessExternalRegistryLabelRemoteRegistryKey: ServerlessExternalRegistryLabelRemoteRegistryVal,
				ServerlessExternalRegistryLabelConfigKey:         ServerlessExternalRegistryLabelConfigVal,
			},
		},
		Type: v1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			"registryAddress": []byte("test-registry-address"),
			"serverAddress":   []byte("test-server-address"),
		},
	}
}
