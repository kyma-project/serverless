package resources

import (
	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	svcTargetPort = intstr.FromInt32(8080)
)

type serviceOptions func(*Service)

// ServiceName - set the service name
func ServiceName(name string) serviceOptions {
	return func(s *Service) {
		s.svcName = name
	}
}

// ServiceAppendSelectorLabels - add additional labels to the service's selector
func ServiceAppendSelectorLabels(labels map[string]string) serviceOptions {
	return func(s *Service) {
		for k, v := range labels {
			s.selectorLabels[k] = v
		}
	}
}

// ServiceTrimClusterInfoLabels - get rid of internal labels like managed-by, function-name or uuid
func ServiceTrimClusterInfoLabels() serviceOptions {
	return func(s *Service) {
		internalLabels := s.function.InternalFunctionLabels()
		for key := range internalLabels {
			delete(s.functionLabels, key)
			delete(s.selectorLabels, key)
		}
	}
}

type Service struct {
	*corev1.Service
	function       *serverlessv1alpha2.Function
	functionLabels map[string]string
	selectorLabels map[string]string
	svcName        string
}

func NewService(f *serverlessv1alpha2.Function, opts ...serviceOptions) *Service {
	s := &Service{
		function:       f,
		functionLabels: f.FunctionLabels(),
		selectorLabels: f.SelectorLabels(),
		svcName:        f.Name,
	}

	for _, o := range opts {
		o(s)
	}

	s.Service = s.construct()
	return s
}

func (s *Service) construct() *corev1.Service {
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.svcName,
			Namespace: s.function.Namespace,
			Labels:    s.functionLabels,
			//TODO: do we need to add annotations here?
			//Annotations: s.functionAnnotations(),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       "http", // it has to be here for istio to work properly
				TargetPort: svcTargetPort,
				Port:       80,
				Protocol:   corev1.ProtocolTCP,
			}},
			Selector: s.selectorLabels,
		},
	}

	return service
}
