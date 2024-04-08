package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *DockerRegistry) IsInState(state State) bool {
	return s.Status.State == state
}

func (s *DockerRegistry) IsCondition(conditionType ConditionType) bool {
	return meta.FindStatusCondition(
		s.Status.Conditions, string(conditionType),
	) != nil
}

func (s *DockerRegistry) IsConditionTrue(conditionType ConditionType) bool {
	condition := meta.FindStatusCondition(s.Status.Conditions, string(conditionType))
	return condition != nil && condition.Status == metav1.ConditionTrue
}

const (
	DefaultEnableInternal   = false
	DefaultRegistryAddress  = "k3d-kyma-registry:5000"
	DefaultServerAddress    = "k3d-kyma-registry:5000"
	EndpointDisabled        = ""
	DefaultEventingEndpoint = "http://eventing-publisher-proxy.kyma-system.svc.cluster.local/publish"
)
