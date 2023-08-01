package tracing

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	tracingOTLPService        = "telemetry-otlp-traces"
	tracingOTLServiceHTTPPort = 4318
)

type eventHandler struct {
}

func (e eventHandler) Create(event event.CreateEvent, q workqueue.RateLimitingInterface) {
	if event.Object == nil {
		return
	}
	svcName := event.Object.GetName()
	if svcName != tracingOTLPService {
		return
	}
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name:      event.Object.GetName(),
		Namespace: event.Object.GetNamespace(),
	}})
}

func (e eventHandler) Update(_ event.UpdateEvent, _ workqueue.RateLimitingInterface) {
	return
}

func (e eventHandler) Delete(event event.DeleteEvent, q workqueue.RateLimitingInterface) {
	if event.Object == nil {
		return
	}
	svcName := event.Object.GetName()
	if svcName != tracingOTLPService {
		return
	}
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name:      event.Object.GetName(),
		Namespace: event.Object.GetNamespace(),
	}})
}

func (e eventHandler) Generic(_ event.GenericEvent, _ workqueue.RateLimitingInterface) {
	return
}

var _ handler.EventHandler = eventHandler{}

func ServiceCollectorWatcher() handler.EventHandler {
	return &eventHandler{}
}

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

	return fmt.Sprintf("%s.%s.svc.cluster.local:%d", svc.Name, svc.Namespace, tracingOTLServiceHTTPPort), nil
}

func findService(name string, svcs *corev1.ServiceList) *corev1.Service {
	for _, svc := range svcs.Items {
		if svc.Name == name {
			return &svc
		}
	}
	return nil
}
