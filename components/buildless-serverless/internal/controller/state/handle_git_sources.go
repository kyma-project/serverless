package state

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing/transport"
	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/fsm"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/git"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
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
				serverlessv1alpha2.ConditionReasonSourceUpdateFailed,
				fmt.Sprintf("Getting git authorization data failed: %s", err.Error()))
			return stopWithError(err)
		}
		m.State.GitAuth = gitAuth
	}

	result := checkLastCommit(m, gitRepository)
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

func checkLastCommit(m *fsm.StateMachine, gitRepository *serverlessv1alpha2.GitRepositorySource) *git.OrderResult {
	url := gitRepository.URL
	ref := gitRepository.Reference
	auth := m.State.GitAuth
	if !m.GitChecker.IsLastCommitCheckOrdered(url, ref, auth) {
		// Order the commit check and return nil result for now
		m.GitChecker.OrderLastCommitCheck(url, ref, auth)
		return nil
	}

	result := m.GitChecker.GetLastCommitCheckResult(url, ref, auth)
	if result != nil && forceGitSourceCheck(m.State.Function) {
		// Clean up the order after getting the result if force check was requested
		m.GitChecker.DeleteLastCommitCheckOrder(url, ref, auth)
	}

	return result
}

func prepareErrorMessage(repoUrl string, err error) string {
	if errors.Is(err, transport.ErrAuthenticationRequired) {
		return fmt.Sprintf("Authentication required for Git repository: %s ", repoUrl)
	}

	return fmt.Sprintf("Git repository: %s source check failed: %s", repoUrl, err.Error())
}

func forceGitSourceCheck(f serverlessv1alpha2.Function) bool {
	if v, ok := f.Annotations[continuousGitCheckoutAnnotation]; ok && strings.ToLower(v) == "true" {
		return true
	}
	return false
}
