package registry

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ServerlessRegistryDefaultSecretName              = "serverless-registry-config-default"
	ServerlessExternalRegistrySecretName             = "serverless-registry-config"
	ServerlessExternalRegistryLabelRemoteRegistryKey = "serverless.kyma-project.io/remote-registry"
	ServerlessExternalRegistryLabelRemoteRegistryVal = "config"
	ServerlessExternalRegistryLabelConfigKey         = "serverless.kyma-project.io/config"
	ServerlessExternalRegistryLabelConfigVal         = "credentials"
	ServerlessRegistryIsInternalKey                  = "isInternal"
	ServerlessDockerRegistryDeploymentName           = "serverless-docker-registry"
	RegistryHTTPEnvKey                               = "REGISTRY_HTTP_SECRET"
)

func ListExternalNamespacedScopeSecrets(ctx context.Context, c client.Client) ([]corev1.Secret, error) {

	// has config label
	remoteRegistryLabelRequirement, _ := labels.NewRequirement(ServerlessExternalRegistryLabelRemoteRegistryKey, selection.Equals, []string{
		ServerlessExternalRegistryLabelRemoteRegistryVal,
	})

	// has not credentials label
	configLabelRequirement, _ := labels.NewRequirement(ServerlessExternalRegistryLabelConfigKey, selection.DoesNotExist, []string{})

	labeledSecrets := corev1.SecretList{}
	err := c.List(ctx, &labeledSecrets, &client.ListOptions{
		LabelSelector: labels.NewSelector().Add(
			*remoteRegistryLabelRequirement,
			*configLabelRequirement,
		),
	})
	if err != nil {
		return nil, err
	}

	secrets := []corev1.Secret{}
	for _, secret := range labeledSecrets.Items {
		if secret.Name == ServerlessExternalRegistrySecretName {
			secrets = append(secrets, secret)
		}
	}

	return secrets, err
}

func GetExternalClusterWideRegistrySecret(ctx context.Context, c client.Client, namespace string) (*corev1.Secret, error) {
	secret := corev1.Secret{}
	key := client.ObjectKey{
		Namespace: namespace,
		Name:      ServerlessExternalRegistrySecretName,
	}
	err := c.Get(ctx, key, &secret)
	if err != nil {
		return nil, client.IgnoreNotFound(err)
	}

	if val, ok := secret.GetLabels()[ServerlessExternalRegistryLabelRemoteRegistryKey]; !ok || val != ServerlessExternalRegistryLabelRemoteRegistryVal {
		return nil, nil
	}
	if val, ok := secret.GetLabels()[ServerlessExternalRegistryLabelConfigKey]; !ok || val != ServerlessExternalRegistryLabelConfigVal {
		return nil, nil
	}

	return &secret, nil
}

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
