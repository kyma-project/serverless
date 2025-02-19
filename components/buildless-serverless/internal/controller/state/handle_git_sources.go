package state

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
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
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:           gitRepository.URL,
		ReferenceName: plumbing.ReferenceName(gitRepository.Repository.Reference),
		SingleBranch:  true,
		Depth:         1,
	})
	if err != nil {
		m.State.Function.UpdateCondition(
			serverlessv1alpha2.ConditionConfigurationReady,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonGitSourceCheckFailed,
			fmt.Sprintf("Git repository: %s clone failed: %s", gitRepository.URL, err.Error()))
		return stop()
	}

	ref, err := r.Head()
	if err != nil {
		m.State.Function.UpdateCondition(
			serverlessv1alpha2.ConditionConfigurationReady,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonGitSourceCheckFailed,
			fmt.Sprintf("Git repository: %s get head failed: %s", gitRepository.URL, err.Error()))
		return stop()
	}

	m.State.Commit = ref.Hash().String()

	return nextState(sFnHandleDeployment)
}
