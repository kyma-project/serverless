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
		errMsgs = append(errMsgs, fmt.Sprintf("name:%s, namespace %s", secret.Name, secret.Namespace))
	}

	return errors.Errorf("Secrets found: %s", strings.Join(errMsgs, ";"))
}
