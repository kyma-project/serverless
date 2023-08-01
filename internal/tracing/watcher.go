package tracing

import (
	telemetryv1alpha1 "github.com/kyma-project/telemetry-manager/apis/telemetry/v1alpha1"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type eventHandler struct {
}

func (e eventHandler) Create(event event.CreateEvent, limitingInterface workqueue.RateLimitingInterface) {
	//TODO implement me
	panic("implement me")
}

func (e eventHandler) Update(event event.UpdateEvent, limitingInterface workqueue.RateLimitingInterface) {
	//TODO implement me
	panic("implement me")
}

func (e eventHandler) Delete(event event.DeleteEvent, limitingInterface workqueue.RateLimitingInterface) {
	//TODO implement me
	panic("implement me")
}

func (e eventHandler) Generic(event event.GenericEvent, limitingInterface workqueue.RateLimitingInterface) {
	//TODO implement me
	panic("implement me")
}

var _ handler.EventHandler = eventHandler{}

func GetSource() source.Source {
	return &source.Kind{Type: &telemetryv1alpha1.TracePipeline{}}
}

func EventWatcher() handler.EventHandler {
	return nil
}

//
//func ConfigureWatcher() {
//	gvkExternal := schema.GroupVersionKind{
//		Group:   "some.group.io",
//		Version: "v1",
//		Kind:    "External",
//	}
//
//	restClient, err := apiutil.RESTClientForGVK(gvkExternal, false, mgr.GetConfig(), serializer.NewCodecFactory(mgr.GetScheme()))
//	if err != nil {
//		setupLog.Error(err, "unable to create REST client")
//	}
//}
