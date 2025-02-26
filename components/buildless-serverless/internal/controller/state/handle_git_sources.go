package state

import (
	"context"
	"fmt"
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/internal/config"
	"github.com/kyma-project/serverless/internal/controller/fsm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"
	"time"
)

const (
	continuousGitCheckoutAnnotation = "serverless.kyma-project.io/continuousGitCheckout"
)

func sFnHandleGitSources(_ context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	if !m.State.Function.HasGitSources() {
		return nextState(sFnHandleDeployment)
	}

	gitRepository := m.State.Function.Spec.Source.GitRepository

	if skipGitSourceCheck(m.State.Function, m.FunctionConfig) {
		m.Log.Info(fmt.Sprintf("skipping function [%s] source check", m.State.Function.Name))
		return nextState(sFnHandleDeployment)
	}

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

func skipGitSourceCheck(f serverlessv1alpha2.Function, cfg config.FunctionConfig) bool {
	if v, ok := f.Annotations[continuousGitCheckoutAnnotation]; ok && strings.ToLower(v) == "true" {
		return false
	}

	// ConditionConfigurationReady is set to true for git functions when the source is updated.
	// if not, this is a new function, we need to do git check.
	configured := f.Status.Condition(serverlessv1alpha2.ConditionConfigurationReady)
	if configured == nil || configured.Status != "True" {
		return false
	}

	return time.Since(configured.LastTransitionTime.Time) < cfg.FunctionReadyRequeueDuration
}
