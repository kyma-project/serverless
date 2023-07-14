package registry

import (
	"context"
	"fmt"
	"github.com/go-errors/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

const (
	ServerlessExternalRegistrySecretName             = "serverless-registry-config"
	ServerlessExternalRegistryLabelRemoteRegistryKey = "serverless.kyma-project.io/remote-registry"
	ServerlessExternalRegistryLabelRemoteRegistryVal = "config"
	ServerlessExternalRegistryLabelConfigKey         = "serverless.kyma-project.io/config"
	ServerlessExternalRegistryLabelConfigVal         = "credentials"
)

func DetectExternalRegistrySecrets(ctx context.Context, c client.Client) error {
	secrets := corev1.SecretList{}
	err := c.List(ctx, &secrets, client.MatchingLabels{ServerlessExternalRegistryLabelRemoteRegistryKey: ServerlessExternalRegistryLabelRemoteRegistryVal})
	if err != nil {
		return err
	}
	if len(secrets.Items) == 0 {
		return nil
	}

	var errMsgs []string
	for _, secret := range secrets.Items {
		errMsgs = append(errMsgs, fmt.Sprintf("found %s/%s secret", secret.Namespace, secret.Name))
	}

	return errors.Errorf("additional registry configuration detected: %s", strings.Join(errMsgs, "; "))
}

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
