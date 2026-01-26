package state

import (
	"context"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/fsm"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnAdjustStatus(_ context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	s := &m.State.Function.Status
	f := m.State.Function
	s.Runtime = f.Spec.Runtime.SupportedRuntimeEquivalent()
	s.RuntimeImage = m.State.BuiltDeployment.RuntimeImage()
	s.Replicas = m.State.ClusterDeployment.Status.Replicas

	// set scale sub-resource
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{MatchLabels: f.SelectorLabels()})
	if err != nil {
		m.Log.Warnf("failed to get selector for labelSelector: %w", err)
		return stopWithError(errors.Wrap(err, "while getting selectors"))
	}
	s.PodSelector = selector.String()

	s.FunctionResourceProfile = m.State.BuiltDeployment.ResourceProfile()
	s.ContainerSecurityContext = m.State.BuiltDeployment.ContainerSecurityContext()
	s.PodSecurityContext = m.State.BuiltDeployment.PodSecurityContext()

	if m.State.Function.HasGitSources() {
		s.GitRepository = &serverlessv1alpha2.GitRepositoryStatus{
			URL: f.Spec.Source.GitRepository.URL,
			Repository: serverlessv1alpha2.Repository{
				BaseDir:   f.Spec.Source.GitRepository.BaseDir,
				Reference: f.Spec.Source.GitRepository.Reference,
			},
			Commit: m.State.Commit,
		}
		s.Repository.BaseDir = f.Spec.Source.GitRepository.BaseDir
		s.Repository.Reference = f.Spec.Source.GitRepository.Reference
		s.Commit = m.State.Commit
	} else {
		s.GitRepository = nil
		s.Repository = serverlessv1alpha2.Repository{}
		s.Commit = ""
	}

	return requeueAfter(m.FunctionConfig.FunctionReadyRequeueDuration)
}
