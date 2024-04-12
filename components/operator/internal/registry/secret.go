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
	RegistryDefaultSecretName              = "serverless-registry-config-default"
	ExternalRegistrySecretName             = "serverless-registry-config"
	ExternalRegistryLabelRemoteRegistryKey = "serverless.kyma-project.io/remote-registry"
	ExternalRegistryLabelRemoteRegistryVal = "config"
	ExternalRegistryLabelConfigKey         = "serverless.kyma-project.io/config"
	ExternalRegistryLabelConfigVal         = "credentials"
	RegistryIsInternalKey                  = "isInternal"
	DockerRegistryDeploymentName           = "internal-docker-registry"
	RegistryHTTPEnvKey                     = "REGISTRY_HTTP_SECRET"
)

func ListExternalNamespacedScopeSecrets(ctx context.Context, c client.Client) ([]corev1.Secret, error) {

	// has config label
	remoteRegistryLabelRequirement, _ := labels.NewRequirement(ExternalRegistryLabelRemoteRegistryKey, selection.Equals, []string{
		ExternalRegistryLabelRemoteRegistryVal,
	})

	// has not credentials label
	configLabelRequirement, _ := labels.NewRequirement(ExternalRegistryLabelConfigKey, selection.DoesNotExist, []string{})

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
		if secret.Name == ExternalRegistrySecretName {
			secrets = append(secrets, secret)
		}
	}

	return secrets, err
}

func GetExternalClusterWideRegistrySecret(ctx context.Context, c client.Client, namespace string) (*corev1.Secret, error) {
	secret := corev1.Secret{}
	key := client.ObjectKey{
		Namespace: namespace,
		Name:      ExternalRegistrySecretName,
	}
	err := c.Get(ctx, key, &secret)
	if err != nil {
		return nil, client.IgnoreNotFound(err)
	}

	if val, ok := secret.GetLabels()[ExternalRegistryLabelRemoteRegistryKey]; !ok || val != ExternalRegistryLabelRemoteRegistryVal {
		return nil, nil
	}
	if val, ok := secret.GetLabels()[ExternalRegistryLabelConfigKey]; !ok || val != ExternalRegistryLabelConfigVal {
		return nil, nil
	}

	return &secret, nil
}

func GetDockerRegistryInternalRegistrySecret(ctx context.Context, c client.Client, namespace string) (*corev1.Secret, error) {
	secret := corev1.Secret{}
	key := client.ObjectKey{
		Namespace: namespace,
		Name:      RegistryDefaultSecretName,
	}
	err := c.Get(ctx, key, &secret)
	if err != nil {
		return nil, client.IgnoreNotFound(err)
	}

	if val, ok := secret.GetLabels()[ExternalRegistryLabelConfigKey]; !ok || val != ExternalRegistryLabelConfigVal {
		return nil, nil
	}

	if val := string(secret.Data[RegistryIsInternalKey]); val != "true" {
		return nil, nil
	}

	return &secret, nil
}

func GetRegistryHTTPSecretEnvValue(ctx context.Context, c client.Client, namespace string) (string, error) {
	deployment := appsv1.Deployment{}
	key := client.ObjectKey{
		Namespace: namespace,
		Name:      DockerRegistryDeploymentName,
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
