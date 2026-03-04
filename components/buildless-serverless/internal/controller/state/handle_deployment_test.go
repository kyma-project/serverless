package state

import (
	"context"
	"errors"
	"testing"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/config"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/fsm"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/resources"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func Test_sFnHandleDeployment(t *testing.T) {
	t.Run("when deployment does not exist on kubernetes should create deployment and apply it", func(t *testing.T) {
		// Arrange
		// some deployment on k8s, but it is not the deployment we expect
		someDeployment := appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sleepy-sammet-name",
				Namespace: "frosty-wilson-ns"}}
		// scheme and fake client
		scheme := runtime.NewScheme()
		require.NoError(t, serverlessv1alpha2.AddToScheme(scheme))
		require.NoError(t, appsv1.AddToScheme(scheme))
		updateWasCalled := false
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&someDeployment).WithInterceptorFuncs(interceptor.Funcs{
			Update: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
				updateWasCalled = true
				return nil
			},
		}).Build()
		// machine with our function
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Function: serverlessv1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "peaceful-merkle-name",
						Namespace: "gifted-khorana-ns"},
					Spec: serverlessv1alpha2.FunctionSpec{
						Runtime: serverlessv1alpha2.NodeJs24,
						Source: serverlessv1alpha2.Source{
							Inline: &serverlessv1alpha2.InlineSource{
								Source: "silly-kowalevski"}},
						Annotations: map[string]string{"mclaren": "inspiring"}}}},
			Log:    zap.NewNop().Sugar(),
			Client: k8sClient,
			Scheme: scheme}

		// Act
		next, result, err := sFnHandleDeployment(context.Background(), &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// we expect stop and requeue
		require.NotNil(t, result)
		require.Equal(t, ctrl.Result{Requeue: true}, *result)
		// no next state (we will stop)
		require.Nil(t, next)
		// deployment has not been updated
		require.False(t, updateWasCalled)
		// function has proper condition
		requireContainsConditionWithMessagePattern(t, m.State.Function.Status,
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionUnknown,
			serverlessv1alpha2.ConditionReasonDeploymentCreated,
			"^Deployment peaceful-merkle-name-\\w+ created$")
		// deployment has been applied to k8s
		clusterDeployments := &appsv1.DeploymentList{}
		getErr := k8sClient.List(context.Background(), clusterDeployments, client.InNamespace("gifted-khorana-ns"))
		require.NoError(t, getErr)
		require.Len(t, clusterDeployments.Items, 1)
		appliedDeployment := clusterDeployments.Items[0]
		require.Regexp(t, "^peaceful-merkle-name-\\w+$", appliedDeployment.Name)
		// deployment should have owner ref to our function
		require.NotEmpty(t, appliedDeployment.OwnerReferences)
		require.Equal(t, "Function", appliedDeployment.OwnerReferences[0].Kind)
		require.Equal(t, "peaceful-merkle-name", appliedDeployment.OwnerReferences[0].Name)
		// function status should be updated with annotations
		require.Equal(t, map[string]string{"mclaren": "inspiring"}, m.State.Function.Status.FunctionAnnotations)
	})
	t.Run("when cannot get deployment from kubernetes should stop processing", func(t *testing.T) {
		// Arrange
		// scheme and fake client
		scheme := runtime.NewScheme()
		require.NoError(t, serverlessv1alpha2.AddToScheme(scheme))
		require.NoError(t, appsv1.AddToScheme(scheme))
		createOrUpdateWasCalled := false
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
			List: func(ctx context.Context, client client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
				return errors.New("magical-hellman error message")
			},
			Get: func(ctx context.Context, client client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
				return errors.New("magical-hellman error message")
			},
			Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
				createOrUpdateWasCalled = true
				return nil
			},
			Update: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
				createOrUpdateWasCalled = true
				return nil
			},
		}).Build()
		// machine with our function
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Function: serverlessv1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nice-matsumoto-name",
						Namespace: "festive-dewdney-ns"},
					Spec: serverlessv1alpha2.FunctionSpec{
						Runtime: serverlessv1alpha2.NodeJs24,
						Source: serverlessv1alpha2.Source{
							Inline: &serverlessv1alpha2.InlineSource{
								Source: "xenodochial-napier"}}}}},
			Log:    zap.NewNop().Sugar(),
			Client: k8sClient,
			Scheme: scheme}

		// Act
		next, result, err := sFnHandleDeployment(context.Background(), &m)

		// Assert
		// we expect error
		require.NotNil(t, err)
		require.ErrorContains(t, err, "magical-hellman error message")
		// no result because of error
		require.Nil(t, result)
		// no next state (we will stop)
		require.Nil(t, next)
		// TODO: should we set condition in this case?
		// deployment has not been created or updated
		require.False(t, createOrUpdateWasCalled)
	})
	t.Run("when deployment does not exist on kubernetes and create fails should stop processing", func(t *testing.T) {
		// Arrange
		// scheme and fake client
		scheme := runtime.NewScheme()
		require.NoError(t, serverlessv1alpha2.AddToScheme(scheme))
		require.NoError(t, appsv1.AddToScheme(scheme))
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
			Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
				return errors.New("competent-goldwasser error message")
			},
		}).Build()
		// machine with our function
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Function: serverlessv1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nostalgic-hugle-name",
						Namespace: "amazing-khayyam-ns"},
					Spec: serverlessv1alpha2.FunctionSpec{
						Runtime: serverlessv1alpha2.NodeJs24,
						Source: serverlessv1alpha2.Source{
							Inline: &serverlessv1alpha2.InlineSource{
								Source: "sleepy-stonebraker"}}}}},
			Log:    zap.NewNop().Sugar(),
			Client: k8sClient,
			Scheme: scheme}

		// Act
		next, result, err := sFnHandleDeployment(context.Background(), &m)

		// Assert
		// we expect error
		require.NotNil(t, err)
		require.ErrorContains(t, err, "competent-goldwasser error message")
		// no result because of error
		require.Nil(t, result)
		// no next state (we will stop)
		require.Nil(t, next)
		// function has proper condition
		requireContainsCondition(t, m.State.Function.Status,
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonDeploymentFailed,
			"Deployment nostalgic-hugle-name-* create failed: competent-goldwasser error message")
	})
	t.Run("when deployment exists on kubernetes but we do not need changes should keep it without changes and go to the next state", func(t *testing.T) {
		// Arrange
		f := serverlessv1alpha2.Function{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "awesome-kapitsa-name",
				Namespace: "stoic-swanson-ns"},
			Spec: serverlessv1alpha2.FunctionSpec{
				Runtime: serverlessv1alpha2.NodeJs24,
				Source: serverlessv1alpha2.Source{
					Inline: &serverlessv1alpha2.InlineSource{
						Source: "affectionate-mclean"}}}}
		fc := config.FunctionConfig{
			Images: config.ImagesConfig{NodeJs24: "boring-bartik"},
		}
		cd := appsv1.Deployment{}
		// identical deployment will be generated inside sFnHandleDeployment
		deployment := resources.NewDeployment(&f, &fc, &cd, "test-commit", nil, "", true).Deployment
		// scheme and fake client
		scheme := runtime.NewScheme()
		require.NoError(t, serverlessv1alpha2.AddToScheme(scheme))
		require.NoError(t, appsv1.AddToScheme(scheme))
		createOrUpdateWasCalled := false
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(deployment).WithInterceptorFuncs(interceptor.Funcs{
			Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
				createOrUpdateWasCalled = true
				return nil
			},
			Update: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
				createOrUpdateWasCalled = true
				return nil
			},
		}).Build()
		// machine with our function
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Function: f},
			FunctionConfig: fc,
			Log:            zap.NewNop().Sugar(),
			Client:         k8sClient,
			Scheme:         scheme}

		// Act
		next, result, err := sFnHandleDeployment(context.Background(), &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// without stopping processing
		require.Nil(t, result)
		// with expected next state
		require.NotNil(t, next)
		requireEqualFunc(t, sFnHandleService, next)
		// deployment has not been created or updated
		require.False(t, createOrUpdateWasCalled)
		// function conditions remain unchanged
		require.Empty(t, m.State.Function.Status.Conditions)
		// fsm stores the generated deployment for next states
		require.NotNil(t, m.State.BuiltDeployment)
		require.Contains(t, m.State.BuiltDeployment.Deployment.Spec.Template.Spec.Containers[0].Env,
			corev1.EnvVar{Name: "FUNC_HANDLER_SOURCE", Value: "affectionate-mclean"})
		require.Equal(t, "boring-bartik", m.State.BuiltDeployment.Deployment.Spec.Template.Spec.Containers[0].Image)
	})
	t.Run("when deployment exists on kubernetes and we need changes should update it and requeue", func(t *testing.T) {
		// Arrange
		// our function
		f := serverlessv1alpha2.Function{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "inspiring-haibt-name",
				Namespace: "heuristic-dubinsky-ns"},
			Spec: serverlessv1alpha2.FunctionSpec{
				Runtime: serverlessv1alpha2.Python312,
				Source: serverlessv1alpha2.Source{
					Inline: &serverlessv1alpha2.InlineSource{
						Source: "brave-euclid"}},
				Annotations: map[string]string{"torvalds": "lucid"}}}
		// deployment which will be returned from kubernetes - empty so there will be a difference
		deployment := appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "inspiring-haibt-name",
				Namespace: "heuristic-dubinsky-ns",
				Labels:    f.InternalFunctionLabels()}}
		// scheme and fake client
		scheme := runtime.NewScheme()
		require.NoError(t, serverlessv1alpha2.AddToScheme(scheme))
		require.NoError(t, appsv1.AddToScheme(scheme))
		createWasCalled := false
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&deployment).WithInterceptorFuncs(interceptor.Funcs{
			Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
				createWasCalled = true
				return nil
			},
		}).Build()
		// machine with our function
		m := fsm.StateMachine{
			State: fsm.SystemState{Function: f},
			FunctionConfig: config.FunctionConfig{
				Images: config.ImagesConfig{Python312: "flamboyant-chatelet"}},
			Log:    zap.NewNop().Sugar(),
			Client: k8sClient,
			Scheme: scheme}

		// Act
		next, result, err := sFnHandleDeployment(context.Background(), &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// we expect stop and requeue
		require.NotNil(t, result)
		require.Equal(t, ctrl.Result{Requeue: true}, *result)
		// no next state (we will stop)
		require.Nil(t, next)
		// function has proper condition
		requireContainsCondition(t, m.State.Function.Status,
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionUnknown,
			serverlessv1alpha2.ConditionReasonDeploymentUpdated,
			"Deployment inspiring-haibt-name updated")
		// deployment has not been created (updated only)
		require.False(t, createWasCalled)
		// deployment has been updated on k8s
		updatedDeployment := &appsv1.Deployment{}
		getErr := k8sClient.Get(context.Background(), client.ObjectKey{
			Name:      "inspiring-haibt-name",
			Namespace: "heuristic-dubinsky-ns",
		}, updatedDeployment)
		require.NoError(t, getErr)
		// deployment should have updated some specific fields
		require.Contains(t, updatedDeployment.Spec.Template.Spec.Containers[0].Env,
			corev1.EnvVar{Name: "FUNC_HANDLER_SOURCE", Value: "brave-euclid"})
		require.Equal(t, "flamboyant-chatelet", updatedDeployment.Spec.Template.Spec.Containers[0].Image)
		// function status should be updated with annotations
		require.Equal(t, map[string]string{"torvalds": "lucid"}, m.State.Function.Status.FunctionAnnotations)
	})
	t.Run("when deployment exists on kubernetes and update fails should stop processing", func(t *testing.T) {
		// Arrange
		// our function
		f := serverlessv1alpha2.Function{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "affectionate-shockley-name",
				Namespace: "boring-swirles-ns"},
			Spec: serverlessv1alpha2.FunctionSpec{
				Runtime: serverlessv1alpha2.Python312,
				Source: serverlessv1alpha2.Source{
					Inline: &serverlessv1alpha2.InlineSource{
						Source: "eager-ardinghelli"}}}}
		// deployment which will be returned from kubernetes - empty so there will be a difference
		deployment := appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "affectionate-shockley-name",
				Namespace: "boring-swirles-ns",
				Labels:    f.InternalFunctionLabels()}}
		// scheme and fake client
		scheme := runtime.NewScheme()
		require.NoError(t, serverlessv1alpha2.AddToScheme(scheme))
		require.NoError(t, appsv1.AddToScheme(scheme))
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&deployment).WithInterceptorFuncs(interceptor.Funcs{
			Update: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
				return errors.New("happy-pare error message")
			},
		}).Build()
		// machine with our function
		m := fsm.StateMachine{
			State: fsm.SystemState{Function: f},
			FunctionConfig: config.FunctionConfig{
				Images: config.ImagesConfig{Python312: "naughty-herschel"}},
			Log:    zap.NewNop().Sugar(),
			Client: k8sClient,
			Scheme: scheme}

		// Act
		next, result, err := sFnHandleDeployment(context.Background(), &m)

		// Assert
		// we expect error
		require.NotNil(t, err)
		require.ErrorContains(t, err, "happy-pare error message")
		// no result because of error
		require.Nil(t, result)
		// no next state (we will stop)
		require.Nil(t, next)
		// function has proper condition
		requireContainsCondition(t, m.State.Function.Status,
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonDeploymentFailed,
			"Deployment affectionate-shockley-name update failed: happy-pare error message")
	})
}

func Test_deploymentChanged(t *testing.T) {
	type args struct {
		a *appsv1.Deployment
		b *appsv1.Deployment
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "when all compared fields are equal should return false",
			args: args{
				a: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Replicas: ptr.To[int32](765),
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"mahavira": "stoic",
									"franklin": "heuristic"},
								Annotations: map[string]string{
									"mahavira": "stoic",
									"franklin": "heuristic"}},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Image:      "stoic-mahavira",
									Command:    []string{"stoic-mahavira", "heuristic-franklin"},
									WorkingDir: "stoic-mahavira",
									Ports: []corev1.ContainerPort{{
										Name:          "stoic-mahavira",
										HostPort:      765,
										ContainerPort: 765,
										Protocol:      "stoic-mahavira",
										HostIP:        "stoic-mahavira"}},
									Env: []corev1.EnvVar{
										{Name: "mahavira", Value: "stoic"},
										{Name: "franklin", Value: "heuristic"}},
									Resources: corev1.ResourceRequirements{
										Limits: corev1.ResourceList{
											"cpu":    resource.MustParse("765m"),
											"memory": resource.MustParse("756Mi")},
										Requests: corev1.ResourceList{
											"cpu":    resource.MustParse("432m"),
											"memory": resource.MustParse("432Mi")}},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:        "stoic-mahavira",
											ReadOnly:    true,
											MountPath:   "stoic-mahavira",
											SubPath:     "stoic-mahavira",
											SubPathExpr: "stoic-mahavira",
										},
										{
											Name:        "heuristic-franklin",
											ReadOnly:    false,
											MountPath:   "heuristic-franklin",
											SubPath:     "heuristic-franklin",
											SubPathExpr: "heuristic-franklin",
										}}}}}}}},
				b: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Replicas: ptr.To[int32](765),
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"franklin": "heuristic",
									"mahavira": "stoic"},
								Annotations: map[string]string{
									"franklin": "heuristic",
									"mahavira": "stoic"}},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Image:      "stoic-mahavira",
									Command:    []string{"stoic-mahavira", "heuristic-franklin"},
									WorkingDir: "stoic-mahavira",
									Ports: []corev1.ContainerPort{{
										Name:          "stoic-mahavira",
										HostPort:      765,
										ContainerPort: 765,
										Protocol:      "stoic-mahavira",
										HostIP:        "stoic-mahavira"}},
									Env: []corev1.EnvVar{
										{Name: "mahavira", Value: "stoic"},
										{Name: "franklin", Value: "heuristic"}},
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											"memory": resource.MustParse("432Mi"),
											"cpu":    resource.MustParse("432m")},
										Limits: corev1.ResourceList{
											"cpu":    resource.MustParse("765m"),
											"memory": resource.MustParse("756Mi")}},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:        "stoic-mahavira",
											ReadOnly:    true,
											MountPath:   "stoic-mahavira",
											SubPath:     "stoic-mahavira",
											SubPathExpr: "stoic-mahavira",
										},
										{
											Name:        "heuristic-franklin",
											ReadOnly:    false,
											MountPath:   "heuristic-franklin",
											SubPath:     "heuristic-franklin",
											SubPathExpr: "heuristic-franklin",
										}}}}}}}},
			},
			want: false,
		},
		{
			name: "when there are no containers should return true",
			args: args{
				a: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{}}}}},
				b: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{}}}}},
			},
			want: true,
		},
		{
			name: "when securityContexts are different should return true",
			args: args{
				a: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									SecurityContext: &corev1.SecurityContext{
										Privileged: ptr.To(true),
									}}}}}}},
				b: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									SecurityContext: &corev1.SecurityContext{
										Privileged: ptr.To(false),
									}}}}}}},
			},
			want: true,
		},
		{
			name: "when podSecurityContexts are different should return true",
			args: args{
				a: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								SecurityContext: &corev1.PodSecurityContext{
									RunAsUser: ptr.To[int64](1000),
									FSGroup:   ptr.To[int64](1000),
								},
							}}}},
				b: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								SecurityContext: &corev1.PodSecurityContext{
									RunAsUser: ptr.To[int64](2000),
								},
							}}}},
			},
			want: true,
		},
		{
			name: "when images are different should return true",
			args: args{
				a: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Image: "dreamy-davinci"}}}}}},
				b: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Image: "crazy-lichterman"}}}}}},
			},
			want: true,
		},
		{
			name: "when labels are different should return true",
			args: args{
				a: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"hofstadter": "laughing"}},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{}}}}}},
				b: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"mcnulty": "compassionate"}},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{}}}}}},
			},
			want: true,
		},
		{
			name: "when annotations are different should return true",
			args: args{
				a: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Annotations: map[string]string{
									"sutherland": "epic"}},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{}}}}}},
				b: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Annotations: map[string]string{
									"volhard": "quizzical"}},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{}}}}}},
			},
			want: true,
		},
		{
			name: "when replicas are different should return true",
			args: args{
				a: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Replicas: ptr.To[int32](979),
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{}}}}}},
				b: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Replicas: ptr.To[int32](555),
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{}}}}}},
			},
			want: true,
		},
		{
			name: "when one of replicas are nil should return true",
			args: args{
				a: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Replicas: ptr.To[int32](343),
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{}}}}}},
				b: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{}}}}}},
			},
			want: true,
		},
		{
			name: "when workingDir are different should return true",
			args: args{
				a: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									WorkingDir: "silly-merkle"}}}}}},
				b: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									WorkingDir: "ecstatic-mendeleev"}}}}}},
			},
			want: true,
		},
		{
			name: "when command are different should return true",
			args: args{
				a: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Command: []string{"ecstatic-jemison"}}}}}}},
				b: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Command: []string{"ecstatic-jemison", "suspicious-gauss"}}}}}}},
			},
			want: true,
		},
		{
			name: "when resources are different should return true",
			args: args{
				a: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											"memory": resource.MustParse("432Mi"),
											"cpu":    resource.MustParse("432m")},
										Limits: corev1.ResourceList{
											"cpu":    resource.MustParse("765m"),
											"memory": resource.MustParse("756Mi")}}}}}}}},
				b: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											"memory": resource.MustParse("432Mi")},
										Limits: corev1.ResourceList{
											"memory": resource.MustParse("756Mi")}}}}}}}},
			},
			want: true,
		},
		{
			name: "when env are different should return true",
			args: args{
				a: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Env: []corev1.EnvVar{
										{Name: "shtern", Value: "agitated"}}}}}}}},
				b: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Env: []corev1.EnvVar{
										{Name: "satoshi", Value: "unruffled"}}}}}}}},
			},
			want: true,
		},
		{
			name: "when volumeMounts are different should return true",
			args: args{
				a: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									VolumeMounts: []corev1.VolumeMount{{
										Name:        "tender-hypatia",
										ReadOnly:    true,
										MountPath:   "tender-hypatia",
										SubPath:     "tender-hypatia",
										SubPathExpr: "tender-hypatia",
									}}}}}}}},
				b: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									VolumeMounts: []corev1.VolumeMount{{
										Name:        "tender-hypatia",
										ReadOnly:    true,
										MountPath:   "hopeful-greider",
										SubPath:     "tender-hypatia",
										SubPathExpr: "tender-hypatia",
									}}}}}}}},
			},
			want: true,
		},
		{
			name: "when ports are different should return true",
			args: args{
				a: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Ports: []corev1.ContainerPort{{
										Name:          "confident-ganguly",
										HostPort:      757,
										ContainerPort: 757,
										Protocol:      "confident-ganguly",
										HostIP:        "confident-ganguly"}}}}}}}},
				b: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Ports: []corev1.ContainerPort{{
										Name:          "confident-ganguly",
										HostPort:      222,
										ContainerPort: 757,
										Protocol:      "confident-ganguly",
										HostIP:        "confident-ganguly"}}}}}}}},
			},
			want: true,
		},
		{
			name: "when (some) not compared fields are different should return false",
			args: args{
				a: &appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "dazzling-tharp",
						APIVersion: "dazzling-tharp"},
					ObjectMeta: metav1.ObjectMeta{
						Name:         "dazzling-tharp",
						GenerateName: "dazzling-tharp",
						Namespace:    "dazzling-tharp",
						UID:          "dazzling-tharp",
						Labels:       map[string]string{"tharp": "dazzling"},
						Annotations:  map[string]string{"tharp": "dazzling"}},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"tharp": "dazzling"}},
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Name:         "dazzling-tharp",
								GenerateName: "dazzling-tharp",
								Namespace:    "dazzling-tharp",
								UID:          "dazzling-tharp"},
							Spec: corev1.PodSpec{
								Volumes: []corev1.Volume{{
									Name: "dazzling-tharp",
									VolumeSource: corev1.VolumeSource{
										EmptyDir: &corev1.EmptyDirVolumeSource{
											Medium: corev1.StorageMediumMemory}}}},
								InitContainers: []corev1.Container{{
									Name: "dazzling-tharp"}},
								Containers: []corev1.Container{{
									Name:            "dazzling-tharp",
									Args:            []string{"dazzling-tharp"},
									ImagePullPolicy: corev1.PullAlways,
									SecurityContext: &corev1.SecurityContext{
										RunAsUser:    ptr.To[int64](579),
										RunAsGroup:   ptr.To[int64](579),
										RunAsNonRoot: ptr.To[bool](true)}}},
								NodeName:          "dazzling-tharp",
								Hostname:          "dazzling-tharp",
								Subdomain:         "dazzling-tharp",
								PriorityClassName: "dazzling-tharp",
								Priority:          ptr.To[int32](579)}},
						MinReadySeconds: 579},
					Status: appsv1.DeploymentStatus{
						Replicas: 579,
						Conditions: []appsv1.DeploymentCondition{{
							Type:   "dazzling-tharp",
							Status: "dazzling-tharp",
							Reason: "dazzling-tharp"}}}},
				b: &appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "thirsty-jemison",
						APIVersion: "thirsty-jemison"},
					ObjectMeta: metav1.ObjectMeta{
						Name:         "thirsty-jemison",
						GenerateName: "thirsty-jemison",
						Namespace:    "thirsty-jemison",
						UID:          "thirsty-jemison",
						Labels:       map[string]string{"jemison": "thirsty"},
						Annotations:  map[string]string{"jemison": "thirsty"}},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"jemison": "thirsty"}},
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Name:         "thirsty-jemison",
								GenerateName: "thirsty-jemison",
								Namespace:    "thirsty-jemison",
								UID:          "thirsty-jemison"},
							Spec: corev1.PodSpec{
								Volumes: []corev1.Volume{{
									Name: "thirsty-jemison",
									VolumeSource: corev1.VolumeSource{
										EmptyDir: &corev1.EmptyDirVolumeSource{
											Medium: corev1.StorageMediumHugePages}}}},
								InitContainers: []corev1.Container{{
									Name: "thirsty-jemison"}},
								Containers: []corev1.Container{{
									Name:            "thirsty-jemison",
									Args:            []string{"thirsty-jemison"},
									ImagePullPolicy: corev1.PullNever,
									SecurityContext: &corev1.SecurityContext{
										RunAsUser:    ptr.To[int64](579),
										RunAsGroup:   ptr.To[int64](579),
										RunAsNonRoot: ptr.To[bool](true)}}},
								NodeName:          "thirsty-jemison",
								Hostname:          "thirsty-jemison",
								Subdomain:         "thirsty-jemison",
								PriorityClassName: "thirsty-jemison",
								Priority:          ptr.To[int32](246)}},
						MinReadySeconds: 246},
					Status: appsv1.DeploymentStatus{
						Replicas: 246,
						Conditions: []appsv1.DeploymentCondition{{
							Type:   "thirsty-jemison",
							Status: "thirsty-jemison",
							Reason: "thirsty-jemison"}}}},
			},
			want: false,
		},
		{
			name: "when there are no init containers should return false",
			args: args{
				a: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								InitContainers: []corev1.Container{},
								Containers:     []corev1.Container{{}}}}}},
				b: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								InitContainers: []corev1.Container{},
								Containers:     []corev1.Container{{}}}}}},
			},
			want: false,
		},
		{
			name: "when init containers count is different should return true",
			args: args{
				a: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								InitContainers: []corev1.Container{{
									Image: "vibrant-booth",
								}},
								Containers: []corev1.Container{{}}}}}},
				b: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								InitContainers: []corev1.Container{},
								Containers:     []corev1.Container{{}}}}}},
			},
			want: true,
		},
		{
			name: "when init container commands are different should return true",
			args: args{
				a: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								InitContainers: []corev1.Container{{
									Image:   "gracious-mclaren",
									Command: []string{"nice-mahavira"},
								}},
								Containers: []corev1.Container{{}}}}}},
				b: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								InitContainers: []corev1.Container{{
									Image:   "gracious-mclaren",
									Command: []string{"modest-kepler"},
								}},
								Containers: []corev1.Container{{}}}}}},
			},
			want: true,
		},
		{
			name: "when init container volume mounts are different should return true",
			args: args{
				a: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								InitContainers: []corev1.Container{{
									Image: "thirsty-matsumoto",
									VolumeMounts: []corev1.VolumeMount{{
										Name:      "name",
										MountPath: "/sweet/carver",
									}}}},
								Containers: []corev1.Container{{}}}}}},
				b: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								InitContainers: []corev1.Container{{
									Image: "thirsty-matsumoto",
									VolumeMounts: []corev1.VolumeMount{{
										Name:      "name",
										MountPath: "/focused/thompson",
									}}}},
								Containers: []corev1.Container{{}}}}}},
			},
			want: true,
		},
		{
			name: "when init container env are different should return true",
			args: args{
				a: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								InitContainers: []corev1.Container{{
									Env: []corev1.EnvVar{
										{Name: "kalam", Value: "admiring"}}}},
								Containers: []corev1.Container{{}}}}}},
				b: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								InitContainers: []corev1.Container{{
									Env: []corev1.EnvVar{
										{Name: "yalow", Value: "nostalgic"}}}},
								Containers: []corev1.Container{{}}}}}},
			},
			want: true,
		},
		{
			name: "when init container workingDir are different should return true",
			args: args{
				a: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								InitContainers: []corev1.Container{{
									WorkingDir: "vibrant-wozniak"}},
								Containers: []corev1.Container{{}}}}}},
				b: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								InitContainers: []corev1.Container{{
									WorkingDir: "romantic-shockley"}},
								Containers: []corev1.Container{{}}}}}},
			},
			want: true,
		},
		{
			name: "when init container images are different should return true",
			args: args{
				a: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								InitContainers: []corev1.Container{{
									Image: "intelligent-galois"}},
								Containers: []corev1.Container{{}}}}}},
				b: &appsv1.Deployment{
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								InitContainers: []corev1.Container{{
									Image: "beautiful-rhodes"}},
								Containers: []corev1.Container{{}}}}}},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := deploymentChanged(tt.args.a, tt.args.b)
			require.Equal(t, tt.want, r)
		})
	}
}
