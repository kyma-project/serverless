package state

import (
	"context"
	"fmt"
	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/fsm"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/git"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"
)

const (
	continuousGitCheckoutAnnotation = "serverless.kyma-project.io/continuousGitCheckout"
)

func sFnHandleGitSources(ctx context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	if !m.State.Function.HasGitSources() {
		return nextState(sFnConfigurationReady)
	}

	gitRepository := m.State.Function.Spec.Source.GitRepository

	if m.State.Function.HasGitAuth() {
		gitAuth, err := git.NewGitAuth(ctx, m.Client, &m.State.Function)
		if err != nil {
			m.State.Function.UpdateCondition(
				serverlessv1alpha2.ConditionConfigurationReady,
				metav1.ConditionFalse,
				serverlessv1alpha2.ConditionReasonGitSourceCheckFailed,
				fmt.Sprintf("Getting git authorization data failed: %s", err.Error()))
			return nil, nil, err
		}
		m.State.GitAuth = gitAuth
	}

	latestCommit, err := m.GitChecker.GetLatestCommit(gitRepository.URL, gitRepository.Reference, m.State.GitAuth, forceGitSourceCheck(m.State.Function))
	if err != nil {
		m.State.Function.UpdateCondition(
			serverlessv1alpha2.ConditionConfigurationReady,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonGitSourceCheckFailed,
			fmt.Sprintf("Git repository: %s source check failed: %s", gitRepository.URL, err.Error()))
		return nil, nil, err
	}

	m.State.Commit = latestCommit

	return nextState(sFnConfigurationReady)
}

func forceGitSourceCheck(f serverlessv1alpha2.Function) bool {
	if v, ok := f.Annotations[continuousGitCheckoutAnnotation]; ok && strings.ToLower(v) == "true" {
		return true
	}
	return false
}
