package state

import (
	"context"
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/internal/controller/fsm"
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
	"testing"
	"time"
)

func Test_sFnHandleDeployment(t *testing.T) {
	//1
	t.Run("when deployment does not exist on kubernetes should create deployment and apply it", func(t *testing.T) {
		// Arrange
		// some deployment on k8s, but it is not the deployment we expect
		someDeployment := appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "peaceful-merkle-name",
				Namespace: "gifted-khorana-ns"}}
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
						Runtime: serverlessv1alpha2.NodeJs22,
						Source: serverlessv1alpha2.Source{
							Inline: &serverlessv1alpha2.InlineSource{
								Source: "silly-kowalevski",
							},
						},
					},
				}},
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
		require.Equal(t, ctrl.Result{RequeueAfter: time.Minute}, *result)
		// no next state (we will stop)
		require.Nil(t, next)
		// deployment has not been updated
		require.False(t, updateWasCalled)
		// function has proper condition
		requireContainsCondition(t, m.State.Function.Status,
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionUnknown,
			serverlessv1alpha2.ConditionReasonDeploymentCreated,
			"Deployment peaceful-merkle-name created")
		// deployment has been applied to k8s
		appliedDeployment := &appsv1.Deployment{}
		getErr := k8sClient.Get(context.Background(), client.ObjectKey{
			Name:      "peaceful-merkle-name",
			Namespace: "gifted-khorana-ns",
		}, appliedDeployment)
		require.NoError(t, getErr)
		// deployment should have owner ref to our function
		require.NotEmpty(t, appliedDeployment.OwnerReferences)
		require.Equal(t, "Function", appliedDeployment.OwnerReferences[0].Kind)
		require.Equal(t, "peaceful-merkle-name", appliedDeployment.OwnerReferences[0].Name)
	})
}

// 1. nie ma deploymentu -> tworzymy go, ustawiamy sukces w conditionie, na k8s jest nasz deployment, mamy requeue
// 2. nie udało się go pobrać deploymentu -> koniec przetwarzania z błędem // czy dodać tu conditiona?
// 3. nie ma deploymentu, błąd tworzenia -> tworzymy go, ale się nie udaje , więc ustawimy condition i kończymy
// 4. jest deployment, nic się nie zmieniło -> nic nie robimy i idziemy do kolejnego stanu; nowy deployment jest w fsm
// 5. jest deployment, mamy zmiany -> ustawiamy condition na sukces i kończymy z requeue
// 6. jest deployment, mamy zmiany, błąd update -> ustawimy condition i kończymy z błędem

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := deploymentChanged(tt.args.a, tt.args.b)
			require.Equal(t, tt.want, r)
		})
	}
}
