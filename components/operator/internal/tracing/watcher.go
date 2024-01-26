package tracing

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	tracingOTLPProtocol       = "http"
	tracingOTLPService        = "telemetry-otlp-traces"
	tracingOTLServiceHTTPPort = 4318
	tracingOTLPPath           = "v1/traces"
)

type eventHandler struct{}

func (e eventHandler) Create(_ context.Context, event event.CreateEvent, q workqueue.RateLimitingInterface) {
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

func (e eventHandler) Update(_ context.Context, _ event.UpdateEvent, _ workqueue.RateLimitingInterface) {
}

func (e eventHandler) Delete(_ context.Context, event event.DeleteEvent, q workqueue.RateLimitingInterface) {
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

func (e eventHandler) Generic(_ context.Context, _ event.GenericEvent, _ workqueue.RateLimitingInterface) {
}

var _ handler.EventHandler = eventHandler{}

func ServiceCollectorWatcher() handler.EventHandler {
	return &eventHandler{}
}
