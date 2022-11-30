package prerequisites

import (
	"context"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	serviceMonitorCRD = "servicemonitors.monitoring.coreos.com"
	virtualServiceCRD = "virtualservices.networking.istio.io"
	gatewayCRD        = "gateways.networking.istio.io"
)

func Check(ctx context.Context, client client.Client, withIstio bool) error {
	crds := []string{}
	crds = append(crds, serviceMonitorCRD)

	if withIstio {
		crds = append(crds, virtualServiceCRD, gatewayCRD)
	}

	return check(ctx, client, crds)
}

func check(ctx context.Context, client client.Client, crds []string) error {
	for i := range crds {
		crd := crds[i]
		if err := getCRD(ctx, client, crd); err != nil {
			return err
		}
	}

	return nil
}

func getCRD(ctx context.Context, client client.Client, name string) error {
	namespacedName := types.NamespacedName{
		Name: name,
	}

	var crd apiextensionsv1.CustomResourceDefinition
	err := client.Get(ctx, namespacedName, &crd)
	return err
}
