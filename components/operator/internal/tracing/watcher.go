package tracing

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

type eventHandler[object client.Object, request reconcile.Request] struct{}

func (e eventHandler[object, request]) Create(_ context.Context, event event.CreateEvent, q workqueue.TypedRateLimitingInterface[request]) {
	if event.Object == nil {
		return
	}
	svcName := event.Object.GetName()
	if svcName != tracingOTLPService {
		return
	}
	q.Add(request{NamespacedName: types.NamespacedName{
		Name:      event.Object.GetName(),
		Namespace: event.Object.GetNamespace(),
	}})
}

func (e eventHandler[object, request]) Update(_ context.Context, _ event.UpdateEvent, _ workqueue.TypedRateLimitingInterface[request]) {
}

func (e eventHandler[object, request]) Delete(_ context.Context, event event.DeleteEvent, q workqueue.TypedRateLimitingInterface[request]) {
	if event.Object == nil {
		return
	}
	svcName := event.Object.GetName()
	if svcName != tracingOTLPService {
		return
	}
	q.Add(request{NamespacedName: types.NamespacedName{
		Name:      event.Object.GetName(),
		Namespace: event.Object.GetNamespace(),
	}})
}

func (e eventHandler[object, request]) Generic(_ context.Context, _ event.GenericEvent, _ workqueue.TypedRateLimitingInterface[request]) {
}

var _ handler.EventHandler = eventHandler[client.Object, reconcile.Request]{}

func ServiceCollectorWatcher() handler.EventHandler {
	return &eventHandler[client.Object, reconcile.Request]{}
}
