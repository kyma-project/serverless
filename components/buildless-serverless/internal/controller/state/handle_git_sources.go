package state

import (
	"context"
	"fmt"
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/internal/controller/fsm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnHandleGitSources(_ context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	if !m.State.Function.HasGitSources() {
		return nextState(sFnHandleDeployment)
	}

	gitRepository := m.State.Function.Spec.Source.GitRepository

	latestCommit, err := m.GitChecker.GetLatestCommit(gitRepository.URL, gitRepository.Reference)
	if err != nil {
		m.State.Function.UpdateCondition(
			serverlessv1alpha2.ConditionConfigurationReady,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonGitSourceCheckFailed,
			fmt.Sprintf("Git repository: %s source check failed: %s", gitRepository.URL, err.Error()))
		return nil, nil, err
	}

	m.State.Commit = latestCommit

	return nextState(sFnHandleDeployment)
}
