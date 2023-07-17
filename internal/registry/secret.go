package registry

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ServerlessExternalRegistrySecretName             = "serverless-registry-config"
	ServerlessExternalRegistryLabelRemoteRegistryKey = "serverless.kyma-project.io/remote-registry"
	ServerlessExternalRegistryLabelRemoteRegistryVal = "config"
	ServerlessExternalRegistryLabelConfigKey         = "serverless.kyma-project.io/config"
	ServerlessExternalRegistryLabelConfigVal         = "credentials"
)

func GetServerlessExternalRegistrySecret(ctx context.Context, c client.Client, namespace string) (*corev1.Secret, error) {
	secret := corev1.Secret{}
	key := client.ObjectKey{
		Namespace: namespace,
		Name:      ServerlessExternalRegistrySecretName,
	}
	err := c.Get(ctx, key, &secret)
	if err != nil {
		return nil, client.IgnoreNotFound(err)
	}

	if secret.Type != corev1.SecretTypeDockerConfigJson {
		return nil, nil
	}
	if val, ok := secret.GetLabels()[ServerlessExternalRegistryLabelRemoteRegistryKey]; !ok || val != ServerlessExternalRegistryLabelRemoteRegistryVal {
		return nil, nil
	}
	if val, ok := secret.GetLabels()[ServerlessExternalRegistryLabelConfigKey]; !ok || val != ServerlessExternalRegistryLabelConfigVal {
		return nil, nil
	}

	return &secret, nil
}
