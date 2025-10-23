package state

import (
	"context"
	"testing"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/config"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/fsm"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/resources"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func Test_sFnAdjustStatus(t *testing.T) {
	t.Run("status is set and requeue after long time from config for inline function", func(t *testing.T) {
		// Arrange
		// machine with our function and previously created/calculated deployment
		f := serverlessv1alpha2.Function{
			ObjectMeta: metav1.ObjectMeta{
				Name: "keen-meitner"},
			Spec: serverlessv1alpha2.FunctionSpec{
				Runtime:              "practical-panini",
				RuntimeImageOverride: "zen-wu",
				Source: serverlessv1alpha2.Source{
					Inline: &serverlessv1alpha2.InlineSource{
						Source: "recursing-bose"}}},
			Status: serverlessv1alpha2.FunctionStatus{}}
		fc := config.FunctionConfig{
			FunctionReadyRequeueDuration: 3546,
			ResourceConfig: config.ResourceConfig{
				Function: config.FunctionResourceConfig{
					Resources: config.Resources{
						DefaultPreset: "charming-dubinsky",
						Presets: config.Preset{
							"charming-dubinsky": config.Resource{}}}}}}
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Function:        f,
				BuiltDeployment: resources.NewDeployment(&f, &fc, nil, "test-commit", nil, ""),
				ClusterDeployment: &appsv1.Deployment{
					Status: appsv1.DeploymentStatus{
						Replicas: int32(686)}}},
			FunctionConfig: fc,
		}

		// Act
		next, result, err := sFnAdjustStatus(context.Background(), &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// we expect stop and requeue
		require.NotNil(t, result)
		require.Equal(t, ctrl.Result{RequeueAfter: 3546}, *result)
		// no next state (we will stop)
		require.Nil(t, next)
		// function should have status image from deployment
		require.Equal(t, serverlessv1alpha2.Runtime("practical-panini"), m.State.Function.Status.Runtime)
		require.Equal(t, "zen-wu", m.State.Function.Status.RuntimeImage)
		require.Equal(t, int32(686), m.State.Function.Status.Replicas)
		require.Contains(t, m.State.Function.Status.PodSelector, "serverless.kyma-project.io/function-name=keen-meitner")
		require.Contains(t, m.State.Function.Status.PodSelector, "serverless.kyma-project.io/managed-by=function-controller")
		require.Contains(t, m.State.Function.Status.PodSelector, "serverless.kyma-project.io/resource=deployment")
		// should be empty beacuse it is inline function
		require.Nil(t, m.State.Function.Status.GitRepository)
		// for backward compatibility
		require.Equal(t, m.State.Function.Status.Repository.BaseDir, "")
		require.Equal(t, m.State.Function.Status.Repository.Reference, "")
		require.Equal(t, m.State.Function.Status.Commit, "")
		// UUID is unset because it is fake object
		require.Contains(t, m.State.Function.Status.PodSelector, "serverless.kyma-project.io/uuid=")
		require.Equal(t, "charming-dubinsky", m.State.Function.Status.FunctionResourceProfile)
	})
	t.Run("status is set and requeue after long time from config for git function", func(t *testing.T) {
		// Arrange
		// machine with our function and previously created/calculated deployment
		f := serverlessv1alpha2.Function{
			ObjectMeta: metav1.ObjectMeta{
				Name: "keen-meitner"},
			Spec: serverlessv1alpha2.FunctionSpec{
				Runtime:              "practical-panini",
				RuntimeImageOverride: "zen-wu",
				Source: serverlessv1alpha2.Source{
					GitRepository: &serverlessv1alpha2.GitRepositorySource{
						URL: "gracious-robinson",
						Repository: serverlessv1alpha2.Repository{
							BaseDir:   "test-base-dir",
							Reference: "test-reference",
						},
					}}},
			Status: serverlessv1alpha2.FunctionStatus{}}
		fc := config.FunctionConfig{
			FunctionReadyRequeueDuration: 3546,
			ResourceConfig: config.ResourceConfig{
				Function: config.FunctionResourceConfig{
					Resources: config.Resources{
						DefaultPreset: "charming-dubinsky",
						Presets: config.Preset{
							"charming-dubinsky": config.Resource{}}}}}}
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Function:        f,
				Commit:          "test-commit",
				BuiltDeployment: resources.NewDeployment(&f, &fc, nil, "test-commit", nil, ""),
				ClusterDeployment: &appsv1.Deployment{
					Status: appsv1.DeploymentStatus{
						Replicas: int32(686)}}},
			FunctionConfig: fc,
		}

		// Act
		next, result, err := sFnAdjustStatus(context.Background(), &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// we expect stop and requeue
		require.NotNil(t, result)
		require.Equal(t, ctrl.Result{RequeueAfter: 3546}, *result)
		// no next state (we will stop)
		require.Nil(t, next)
		// function should have status image from deployment
		require.Equal(t, serverlessv1alpha2.Runtime("practical-panini"), m.State.Function.Status.Runtime)
		require.Equal(t, "zen-wu", m.State.Function.Status.RuntimeImage)
		require.Equal(t, int32(686), m.State.Function.Status.Replicas)
		require.Contains(t, m.State.Function.Status.PodSelector, "serverless.kyma-project.io/function-name=keen-meitner")
		require.Contains(t, m.State.Function.Status.PodSelector, "serverless.kyma-project.io/managed-by=function-controller")
		require.Contains(t, m.State.Function.Status.PodSelector, "serverless.kyma-project.io/resource=deployment")
		// UUID is unset because it is fake object
		require.Contains(t, m.State.Function.Status.PodSelector, "serverless.kyma-project.io/uuid=")
		require.Equal(t, "charming-dubinsky", m.State.Function.Status.FunctionResourceProfile)
		// function should have commit from git, url and reference in status
		require.NotNil(t, m.State.Function.Status.GitRepository)
		require.Equal(t, m.State.Function.Status.GitRepository.Repository.BaseDir, "test-base-dir")
		require.Equal(t, m.State.Function.Status.GitRepository.Repository.Reference, "test-reference")
		require.Equal(t, m.State.Function.Status.GitRepository.Commit, "test-commit")
		require.Equal(t, m.State.Function.Status.GitRepository.URL, "gracious-robinson")
		// for backward compatibility
		require.Equal(t, m.State.Function.Status.Repository.BaseDir, "test-base-dir")
		require.Equal(t, m.State.Function.Status.Repository.Reference, "test-reference")
		require.Equal(t, m.State.Function.Status.Commit, "test-commit")
	})
	t.Run("function resource profile is set to custom when there is resource definition", func(t *testing.T) {
		// Arrange
		// machine with our function and previously created/calculated deployment
		f := serverlessv1alpha2.Function{
			ObjectMeta: metav1.ObjectMeta{
				Name: "determined-banzai"},
			Spec: serverlessv1alpha2.FunctionSpec{
				Runtime: "practical-poincare",
				Source: serverlessv1alpha2.Source{
					Inline: &serverlessv1alpha2.InlineSource{
						Source: "nervous-kilby"}},
				ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
					Function: &serverlessv1alpha2.ResourceRequirements{
						Resources: &corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    k8sresource.MustParse("789m"),
								corev1.ResourceMemory: k8sresource.MustParse("678Mi"),
							},
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    k8sresource.MustParse("345m"),
								corev1.ResourceMemory: k8sresource.MustParse("234Mi"),
							},
						},
					},
				},
			},
			Status: serverlessv1alpha2.FunctionStatus{}}
		fc := config.FunctionConfig{
			FunctionReadyRequeueDuration: 3546,
			ResourceConfig: config.ResourceConfig{
				Function: config.FunctionResourceConfig{
					Resources: config.Resources{
						DefaultPreset: "objective-moore"}}}}
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Function:        f,
				BuiltDeployment: resources.NewDeployment(&f, &fc, nil, "test-commit", nil, ""),
				ClusterDeployment: &appsv1.Deployment{
					Status: appsv1.DeploymentStatus{
						Replicas: int32(686)}}},
			FunctionConfig: fc,
		}

		// Act
		next, result, err := sFnAdjustStatus(context.Background(), &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// we expect stop and requeue
		require.NotNil(t, result)
		require.Equal(t, ctrl.Result{RequeueAfter: 3546}, *result)
		// no next state (we will stop)
		require.Nil(t, next)
		require.Equal(t, "custom", m.State.Function.Status.FunctionResourceProfile)
	})
	t.Run("function resource profile is set to value from profile", func(t *testing.T) {
		// Arrange
		// machine with our function and previously created/calculated deployment
		f := serverlessv1alpha2.Function{
			ObjectMeta: metav1.ObjectMeta{
				Name: "priceless-cohen"},
			Spec: serverlessv1alpha2.FunctionSpec{
				Runtime: "brave-easley",
				Source: serverlessv1alpha2.Source{
					Inline: &serverlessv1alpha2.InlineSource{
						Source: "angry-newton"}},
				ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
					Function: &serverlessv1alpha2.ResourceRequirements{
						Profile: "frosty-aryabhata",
					},
				},
			},
			Status: serverlessv1alpha2.FunctionStatus{}}
		fc := config.FunctionConfig{
			FunctionReadyRequeueDuration: 3546,
			ResourceConfig: config.ResourceConfig{
				Function: config.FunctionResourceConfig{
					Resources: config.Resources{
						DefaultPreset: "zealous-grothendieck",
						Presets: config.Preset{
							"frosty-aryabhata": config.Resource{}}}}}}
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Function:        f,
				BuiltDeployment: resources.NewDeployment(&f, &fc, nil, "test-commit", nil, ""),
				ClusterDeployment: &appsv1.Deployment{
					Status: appsv1.DeploymentStatus{
						Replicas: int32(686)}}},
			FunctionConfig: fc,
		}

		// Act
		next, result, err := sFnAdjustStatus(context.Background(), &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// we expect stop and requeue
		require.NotNil(t, result)
		require.Equal(t, ctrl.Result{RequeueAfter: 3546}, *result)
		// no next state (we will stop)
		require.Nil(t, next)
		require.Equal(t, "frosty-aryabhata", m.State.Function.Status.FunctionResourceProfile)
	})
}
