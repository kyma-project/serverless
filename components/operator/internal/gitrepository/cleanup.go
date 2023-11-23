package gitrepository

import (
	"context"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	gitRepoCRDName = "gitrepositories.serverless.kyma-project.io"
)

// Cleanup removes gitrepository CRD and its resources
func Cleanup(ctx context.Context, c client.Client) error {
	crd, err := getCRD(ctx, c)
	if err != nil {
		return client.IgnoreNotFound(err)
	}

	return c.Delete(ctx, crd, &client.DeleteOptions{})
}

func getCRD(ctx context.Context, client client.Client) (*apiextensionsv1.CustomResourceDefinition, error) {
	var crd apiextensionsv1.CustomResourceDefinition
	err := client.Get(ctx, types.NamespacedName{
		Name: gitRepoCRDName,
	}, &crd)
	return &crd, err
}
