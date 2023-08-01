package tracing

import (
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
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
