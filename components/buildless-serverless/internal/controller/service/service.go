package service

import (
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/internal/config"
	"github.com/kyma-project/serverless/internal/controller/fsm"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	svcTargetPort = intstr.FromInt32(8080)
)

type Service struct {
	*corev1.Service
	functionConfig config.FunctionConfig
	function       *serverlessv1alpha2.Function
}

func New(m *fsm.StateMachine) *Service {
	s := &Service{
		functionConfig: m.FunctionConfig,
		function:       &m.State.Instance,
	}
	s.Service = s.construct()
	return s
}

func (s *Service) construct() *corev1.Service {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.function.Name,
			Namespace: s.function.Namespace,
			//TODO: do we need to add labels or annotations here?
			//Labels:      s.functionLabels(),
			//Annotations: s.functionAnnotations(),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       "http", // it has to be here for istio to work properly
				TargetPort: svcTargetPort,
				Port:       80,
				Protocol:   corev1.ProtocolTCP,
			}},
			Selector: map[string]string{
				// TODO: do we need to add more labels here?
				serverlessv1alpha2.FunctionNameLabel: s.function.GetName(),
			},
		},
	}

	return service
}
