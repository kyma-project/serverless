package registry

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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

func FixServerlessDockerRegistryPVCWithStorage(storage resource.Quantity) *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		TypeMeta: v12.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: v12.ObjectMeta{
			Name:      PvcName,
			Namespace: "kyma-system",
		},
		Spec: v1.PersistentVolumeClaimSpec{
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					"storage": storage,
				},
			},
		},
	}
}
