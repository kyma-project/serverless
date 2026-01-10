package state

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/fsm"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/git/automock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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
		requireEqualFunc(t, sFnConfigurationReady, next)
		// function conditions remain unchanged
		require.Empty(t, m.State.Function.Status.Conditions)
		// no commit change, it should be changed only for git functions
		require.Equal(t, "", m.State.Commit)
	})
	t.Run("for git function where the commit should not be empty and move to the nextState", func(t *testing.T) {
		// Arrange
		// machine with our function
		gitMock := new(automock.LastCommitChecker)
		gitMock.On("GetLatestCommit", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("latest-test-commit", nil)
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
							}}},
					Status: serverlessv1alpha2.FunctionStatus{
						GitRepository: &serverlessv1alpha2.GitRepositoryStatus{
							Commit: "test-commit"}}}},
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
		requireEqualFunc(t, sFnConfigurationReady, next)
		// function conditions changed
		requireContainsCondition(t, m.State.Function.Status,
			serverlessv1alpha2.ConditionConfigurationReady,
			metav1.ConditionTrue,
			serverlessv1alpha2.ConditionReasonSourceUpdated,
			"Function source updated")
		// commit change, it should be changed only for git functions
		require.Equal(t, "latest-test-commit", m.State.Commit)
	})
	t.Run("for git function where the commit should be empty and stop with condition", func(t *testing.T) {
		// Arrange
		// machine with our function
		gitMock := new(automock.LastCommitChecker)
		gitMock.On("GetLatestCommit", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("test-error"))
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
			serverlessv1alpha2.ConditionReasonSourceUpdateFailed,
			"Git repository: test-url source check failed: test-error")
		// no  next state
		require.Nil(t, next)
		// commit did not change
		require.Equal(t, "", m.State.Commit)
		// git auth should be empty
		require.Nil(t, m.State.GitAuth)

	})
	t.Run("do not skip source check for updated function and return commit", func(t *testing.T) {
		// Arrange
		// machine with our function
		gitMock := new(automock.LastCommitChecker)
		gitMock.On("GetLatestCommit", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("latest-commit", nil)
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Function: serverlessv1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "new-function",
						Namespace: "default",
					},
					Spec: serverlessv1alpha2.FunctionSpec{
						Source: serverlessv1alpha2.Source{
							GitRepository: &serverlessv1alpha2.GitRepositorySource{
								URL: "test-url",
								Repository: serverlessv1alpha2.Repository{
									Reference: "main",
								},
							},
						},
					},
					Status: serverlessv1alpha2.FunctionStatus{
						Conditions: []metav1.Condition{
							{
								Type:               string(serverlessv1alpha2.ConditionConfigurationReady),
								Status:             metav1.ConditionTrue,
								LastTransitionTime: metav1.Now(),
							},
						},
						GitRepository: &serverlessv1alpha2.GitRepositoryStatus{
							Commit: "old-commit",
						},
					},
				},
			},
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
		requireEqualFunc(t, sFnConfigurationReady, next)
		// function has proper condition
		requireContainsCondition(t, m.State.Function.Status,
			serverlessv1alpha2.ConditionConfigurationReady,
			metav1.ConditionTrue, serverlessv1alpha2.ConditionReasonSourceUpdated, "Function source updated")
		// commit chang
		require.Equal(t, "latest-commit", m.State.Commit)
	})
	t.Run("for git function with git auth where the commit should not be empty and move to the nextState", func(t *testing.T) {
		// Arrange
		// secret on k8s
		secret := corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "frosty-morse",
				Namespace: "sharp-williams"},
			Data: map[string][]byte{
				"username": []byte("gould"),
				"password": []byte("pensive"),
			}}

		// scheme and fake client
		scheme := runtime.NewScheme()
		require.NoError(t, corev1.AddToScheme(scheme))
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&secret).Build()
		// machine with our function
		gitMock := new(automock.LastCommitChecker)
		gitMock.On("GetLatestCommit", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("latest-test-commit", nil)
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Function: serverlessv1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "lucid-murdock",
						Namespace: "sharp-williams"},
					Spec: serverlessv1alpha2.FunctionSpec{
						Runtime: serverlessv1alpha2.NodeJs22,
						Source: serverlessv1alpha2.Source{
							GitRepository: &serverlessv1alpha2.GitRepositorySource{
								URL: "test-url",
								Repository: serverlessv1alpha2.Repository{
									BaseDir:   "main",
									Reference: "test-reference",
								},
								Auth: &serverlessv1alpha2.RepositoryAuth{
									Type:       serverlessv1alpha2.RepositoryAuthBasic,
									SecretName: "frosty-morse",
								},
							}}},
					Status: serverlessv1alpha2.FunctionStatus{
						GitRepository: &serverlessv1alpha2.GitRepositoryStatus{
							Commit: "test-commit"}},
				}},
			Log:        zap.NewNop().Sugar(),
			Client:     k8sClient,
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
		requireEqualFunc(t, sFnConfigurationReady, next)
		// function conditions changed
		requireContainsCondition(t, m.State.Function.Status,
			serverlessv1alpha2.ConditionConfigurationReady,
			metav1.ConditionTrue,
			serverlessv1alpha2.ConditionReasonSourceUpdated,
			"Function source updated")
		// commit change, it should be changed only for git functions
		require.Equal(t, "latest-test-commit", m.State.Commit)
		// git auth should be set
		require.NotNil(t, m.State.GitAuth)
		authEnvs, _ := json.Marshal(m.State.GitAuth.GetAuthEnvs())
		require.Contains(t, string(authEnvs), "basic")
		require.Contains(t, string(authEnvs), "frosty-morse")
	})
}
