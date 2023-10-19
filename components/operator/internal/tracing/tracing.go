package tracing

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetTraceCollectorURL(ctx context.Context, c client.Client) (string, error) {
	svcs := &corev1.ServiceList{}
	err := c.List(ctx, svcs, &client.ListOptions{})
	if err != nil {
		return "", errors.Wrap(err, "while listing services")
	}
	svc := findService(tracingOTLPService, svcs)
	if svc == nil {
		return "", nil
	}

	return fmt.Sprintf("%s://%s.%s.svc.cluster.local:%d/%s", tracingOTLPProtocol, svc.Name, svc.Namespace, tracingOTLServiceHTTPPort, tracingOTLPPath), nil
}

func findService(name string, svcs *corev1.ServiceList) *corev1.Service {
	for _, svc := range svcs.Items {
		if svc.Name == name {
			return &svc
		}
	}
	return nil
}
