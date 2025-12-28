package legacy

import (
	"context"

	"github.com/kyma-project/manager-toolkit/installation/base/resource"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ConfigLabelKey                = "serverless.kyma-project.io/config"
	DockerfileConfigmapLabelValue = "runtime"
	ServiceAccountLabelValue      = "service-account"
	RegistrySecretLabelValue      = "credentials"
)

// RemoveResourceFromAllNamespaces removes the given resource template from all namespaces in the cluster
func RemoveResourceFromAllNamespaces(ctx context.Context, c client.Client, log *zap.SugaredLogger, template unstructured.Unstructured) (bool, error) {
	done := true
	var namespaces corev1.NamespaceList
	if err := c.List(ctx, &namespaces); err != nil {
		return false, errors.Wrap(err, "couldn't get namespaces during Serverless uninstallation")
	}

	for _, namespace := range namespaces.Items {
		obj := template.DeepCopy()
		obj.SetNamespace(namespace.GetName())
		d, err := resource.Delete(ctx, c, log, *obj)
		if err != nil {
			return false, err
		}
		if !d {
			done = d
		}
	}

	return done, nil
}
