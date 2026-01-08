package state

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-git/go-git/v5/plumbing/transport"
	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/fsm"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/git"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
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
				serverlessv1alpha2.ConditionReasonSourceUpdateFailed,
				fmt.Sprintf("Getting git authorization data failed: %s", err.Error()))
			return stopWithError(err)
		}
		m.State.GitAuth = gitAuth
	}

	orderID := string(m.State.Function.GetUID())
	m.GitChecker.MakeOrder(orderID, gitRepository.URL, gitRepository.Reference, m.State.GitAuth)

	result := m.GitChecker.CollectOrder(orderID)
	if result == nil {
		// Commit check is still in progress, requeue the reconciliation
		return requeueAfter(250 * time.Millisecond)
	}

	if result.Error != nil {
		m.State.Function.UpdateCondition(
			serverlessv1alpha2.ConditionConfigurationReady,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonSourceUpdateFailed,
			prepareErrorMessage(gitRepository.URL, result.Error))
		return stopWithError(result.Error)
	}

	if m.State.Function.Status.GitRepository == nil || m.State.Function.Status.GitRepository.Commit != result.Commit {
		m.State.Function.UpdateCondition(
			serverlessv1alpha2.ConditionConfigurationReady,
			metav1.ConditionTrue,
			serverlessv1alpha2.ConditionReasonSourceUpdated,
			"Function source updated")
	}

	m.State.Commit = result.Commit

	return nextState(sFnConfigurationReady)
}

func prepareErrorMessage(repoUrl string, err error) string {
	if errors.Is(err, transport.ErrAuthenticationRequired) {
		return fmt.Sprintf("Authentication required for Git repository: %s ", repoUrl)
	}

	return fmt.Sprintf("Git repository: %s source check failed: %s", repoUrl, err.Error())
}
