package resources

import (
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	svcTargetPort = intstr.FromInt32(8080)
)

type Service struct {
	*corev1.Service
	function *serverlessv1alpha2.Function
}

func NewService(f *serverlessv1alpha2.Function) *Service {
	s := &Service{
		function: f,
	}
	s.Service = s.construct()
	return s
}

func (s *Service) construct() *corev1.Service {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.function.Name,
			Namespace: s.function.Namespace,
			Labels:    s.function.FunctionLabels(),
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
			Selector: s.function.SelectorLabels(),
		},
	}

	return service
}
