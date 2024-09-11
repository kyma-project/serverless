package registry

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ServerlessRegistryDefaultSecretName      = "serverless-registry-config-default"
	ServerlessExternalRegistryLabelConfigKey = "serverless.kyma-project.io/config"
	ServerlessExternalRegistryLabelConfigVal = "credentials"
	ServerlessRegistryIsInternalKey          = "isInternal"
	ServerlessDockerRegistryDeploymentName   = "serverless-docker-registry"
	RegistryHTTPEnvKey                       = "REGISTRY_HTTP_SECRET"
)

func GetServerlessInternalRegistrySecret(ctx context.Context, c client.Client, namespace string) (*corev1.Secret, error) {
	secret := corev1.Secret{}
	key := client.ObjectKey{
		Namespace: namespace,
		Name:      ServerlessRegistryDefaultSecretName,
	}
	err := c.Get(ctx, key, &secret)
	if err != nil {
		return nil, client.IgnoreNotFound(err)
	}

	if val, ok := secret.GetLabels()[ServerlessExternalRegistryLabelConfigKey]; !ok || val != ServerlessExternalRegistryLabelConfigVal {
		return nil, nil
	}

	if val := string(secret.Data[ServerlessRegistryIsInternalKey]); val != "true" {
		return nil, nil
	}

	return &secret, nil
}

func GetRegistryHTTPSecretEnvValue(ctx context.Context, c client.Client, namespace string) (string, error) {
	deployment := appsv1.Deployment{}
	key := client.ObjectKey{
		Namespace: namespace,
		Name:      ServerlessDockerRegistryDeploymentName,
	}
	err := c.Get(ctx, key, &deployment)
	if err != nil {
		return "", client.IgnoreNotFound(err)
	}

	envs := deployment.Spec.Template.Spec.Containers[0].Env
	for _, v := range envs {
		if v.Name == RegistryHTTPEnvKey && v.Value != "" {
			return v.Value, nil
		}
	}

	return "", nil
}
