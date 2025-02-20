package state

import (
	"context"
	"errors"
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/internal/controller/fsm"
	"github.com/kyma-project/serverless/internal/controller/git/automock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_sFnHandleGitSources(t *testing.T) {
	t.Run("for inline function where the commit should be empty and move to the nextState", func(t *testing.T) {
		// Arrange
		// machine with our function
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Function: serverlessv1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nice-matsumoto-name",
						Namespace: "festive-dewdney-ns"},
					Spec: serverlessv1alpha2.FunctionSpec{
						Runtime: serverlessv1alpha2.NodeJs22,
						Source: serverlessv1alpha2.Source{
							Inline: &serverlessv1alpha2.InlineSource{
								Source: "xenodochial-napier"}}}}},
			Log: zap.NewNop().Sugar(),
		}

		// Act
		next, result, err := sFnHandleGitSources(context.Background(), &m)

		// Assert
		// we are not expecting error
		require.Nil(t, err)
		// no result
		require.Nil(t, result)
		// with expected next state
		require.NotNil(t, next)
		requireEqualFunc(t, sFnHandleDeployment, next)
		// function conditions remain unchanged
		require.Empty(t, m.State.Function.Status.Conditions)
		// no commit change, it should be changed only for git functions
		require.Equal(t, "", m.State.Commit)
	})
	t.Run("for git function where the commit should not be empty and move to the nextState", func(t *testing.T) {
		// Arrange
		// machine with our function
		gitMock := new(automock.LastCommitChecker)
		gitMock.On("GetLatestCommit", mock.Anything, mock.Anything).Return("latest-test-commit", nil)
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Function: serverlessv1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nice-matsumoto-name",
						Namespace: "festive-dewdney-ns"},
					Spec: serverlessv1alpha2.FunctionSpec{
						Runtime: serverlessv1alpha2.NodeJs22,
						Source: serverlessv1alpha2.Source{
							GitRepository: &serverlessv1alpha2.GitRepositorySource{
								URL: "test-url",
								Repository: serverlessv1alpha2.Repository{
									BaseDir:   "main",
									Reference: "test-reference",
								},
							}}}}},
			Log:        zap.NewNop().Sugar(),
			GitChecker: gitMock,
		}

		// Act
		next, result, err := sFnHandleGitSources(context.Background(), &m)

		// Assert
		// we are not expecting error
		require.Nil(t, err)
		// no result
		require.Nil(t, result)
		// with expected next state
		require.NotNil(t, next)
		requireEqualFunc(t, sFnHandleDeployment, next)
		// function conditions remain unchanged
		require.Empty(t, m.State.Function.Status.Conditions)
		// commit change, it should be changed only for git functions
		require.Equal(t, "latest-test-commit", m.State.Commit)
	})
	t.Run("for git function where the commit should be empty and stop with condition", func(t *testing.T) {
		// Arrange
		// machine with our function
		gitMock := new(automock.LastCommitChecker)
		gitMock.On("GetLatestCommit", mock.Anything, mock.Anything).Return("", errors.New("test-error"))
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Function: serverlessv1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nice-matsumoto-name",
						Namespace: "festive-dewdney-ns"},
					Spec: serverlessv1alpha2.FunctionSpec{
						Runtime: serverlessv1alpha2.NodeJs22,
						Source: serverlessv1alpha2.Source{
							GitRepository: &serverlessv1alpha2.GitRepositorySource{
								URL: "test-url",
								Repository: serverlessv1alpha2.Repository{
									BaseDir:   "main",
									Reference: "test-reference",
								},
							}}}}},
			Log:        zap.NewNop().Sugar(),
			GitChecker: gitMock,
		}

		// Act
		next, result, err := sFnHandleGitSources(context.Background(), &m)

		// Assert
		// we are expecting error
		require.NotNil(t, err)
		require.ErrorContains(t, err, "test-error")
		// no result (error)
		require.Nil(t, result)
		// function has proper condition
		requireContainsCondition(t, m.State.Function.Status,
			serverlessv1alpha2.ConditionConfigurationReady,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonGitSourceCheckFailed,
			"Git repository: test-url source check failed: test-error")
		// no  next state
		require.Nil(t, next)
		// commit did not change
		require.Equal(t, "", m.State.Commit)
	})
}
