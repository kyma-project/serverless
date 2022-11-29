package prerequisites

import (
	"context"
	"strings"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	serviceMonitorCRD = "servicemonitors.monitoring.coreos.com"
	virtualServiceCRD = "virtualservices.networking.istio.io"
	gatewayCRD        = "gateways.networking.istio.io"
)

func Check(ctx context.Context, client client.Client, serverless *v1alpha1.Serverless) error {
	if err := isServiceMonitorInstalled(ctx, client); err != nil {
		return err
	}

	if !isIstioNeeded(serverless) {
		return nil
	}

	if err := isVirtualServiceInstalled(ctx, client); err != nil {
		return err
	}

	gateway := serverless.Spec.DockerRegistry.Gateway
	if err := isGatewayInstalled(ctx, client, *gateway); err != nil {
		return err
	}

	return nil
}

func isGatewayInstalled(ctx context.Context, client client.Client, gateway string) error {
	if err := getCRD(ctx, client, gatewayCRD); err != nil {
		return err
	}

	namespacedName := newNamespacedNameFromString(gateway)

	return client.Get(ctx, namespacedName, &v1beta1.Gateway{})
}

func isVirtualServiceInstalled(ctx context.Context, client client.Client) error {
	return getCRD(ctx, client, virtualServiceCRD)
}

func isServiceMonitorInstalled(ctx context.Context, client client.Client) error {
	return getCRD(ctx, client, serviceMonitorCRD)
}

func isIstioNeeded(serverless *v1alpha1.Serverless) bool {
	isNeeded := false
	if serverless.Spec.DockerRegistry != nil &&
		serverless.Spec.DockerRegistry.EnableInternal != nil {
		isNeeded = *serverless.Spec.DockerRegistry.EnableInternal
	}

	return isNeeded
}

func serverlessGateway(serverless *v1alpha1.Serverless) string {
	gateway := ""
	spec := serverless.Spec
	if spec.DockerRegistry != nil &&
		spec.DockerRegistry.Gateway != nil {
		gateway = *spec.DockerRegistry.Gateway
	}

	return gateway
}

func getCRD(ctx context.Context, client client.Client, name string) error {
	namespacedName := types.NamespacedName{
		Name: name,
	}

	var crd v1.CustomResourceDefinition
	return client.Get(ctx, namespacedName, &crd)
}

// TODO: search for this method

const (
	separator = '/'
)

func newNamespacedNameFromString(s string) types.NamespacedName {
	nn := types.NamespacedName{}
	result := strings.Split(s, string(separator))
	if len(result) == 2 {
		nn.Namespace = result[0]
		nn.Name = result[1]
	}
	return nn
}
