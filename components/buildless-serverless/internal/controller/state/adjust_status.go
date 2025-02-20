package state

import (
	"context"
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/internal/config"
	"github.com/kyma-project/serverless/internal/controller/fsm"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnAdjustStatus(_ context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	s := &m.State.Function.Status
	s.Runtime = m.State.Function.Spec.Runtime
	s.RuntimeImage = m.State.BuiltDeployment.RuntimeImage()
	s.Replicas = m.State.ClusterDeployment.Status.Replicas

	// set scale sub-resource
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{MatchLabels: m.State.Function.SelectorLabels()})
	if err != nil {
		m.Log.Warnf("failed to get selector for labelSelector: %w", err)
		return stopWithEventualError(errors.Wrap(err, "while getting selectors"))
	}
	s.PodSelector = selector.String()

	s.FunctionResourceProfile = getUsedResourceFunctionPreset(m.State.Function.Spec.ResourceConfiguration, m.FunctionConfig)
	
	if m.State.Function.HasGitSources() {
		s.Repository.BaseDir = m.State.Function.Spec.Source.GitRepository.BaseDir
		s.Repository.Reference = m.State.Function.Spec.Source.GitRepository.Reference
		s.Commit = m.State.Commit
	} else {
		s.Repository = serverlessv1alpha2.Repository{}
		s.Commit = ""
	}

	return requeueAfter(m.FunctionConfig.FunctionReadyRequeueDuration)
}

func getUsedResourceFunctionPreset(resourceConfiguration *serverlessv1alpha2.ResourceConfiguration, functionConfig config.FunctionConfig) string {
	defaultPreset := functionConfig.ResourceConfig.Function.Resources.DefaultPreset
	if resourceConfiguration == nil || resourceConfiguration.Function == nil {
		return defaultPreset
	}
	resourceRequirements := resourceConfiguration.Function
	if resourceRequirements.Resources != nil {
		return "custom"
	}
	return resourceRequirements.Profile
}
