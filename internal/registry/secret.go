package registry

import (
	"context"
	"fmt"
	"github.com/go-errors/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func DetectExternalRegistrySecrets(ctx context.Context, c client.Client) error {
	secrets := corev1.SecretList{}
	err := c.List(ctx, &secrets, client.MatchingLabels{"serverless.kyma-project.io/remote-registry": "config"})
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

func GetExternalRegistrySecret(ctx context.Context, c client.Client, namespace string) (*corev1.Secret, error) {
	secret := corev1.Secret{}
	key := client.ObjectKey{
		Namespace: namespace,
		Name:      "serverless-registry-config",
	}
	err := c.Get(ctx, key, &secret)
	if err != nil {
		return nil, client.IgnoreNotFound(err)
	}

	if secret.Type != corev1.SecretTypeDockerConfigJson {
		return nil, nil
	}
	if val, ok := secret.GetLabels()["serverless.kyma-project.io/remote-registry"]; !ok || val != "config" {
		return nil, nil
	}
	if val, ok := secret.GetLabels()["serverless.kyma-project.io/config"]; !ok || val != "credentials" {
		return nil, nil
	}

	return &secret, nil
}
