package serverless

import (
	"context"
	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/automock"
	serverlessResource "github.com/kyma-project/kyma/components/function-controller/internal/resource"
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	controllerruntime "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestValidation(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	minResourcesCfg := ResourceConfig{
		Function: FunctionResourceConfig{
			Resources: Resources{
				MinRequestedCPU:    Quantity{resource.MustParse("10m")},
				MinRequestedMemory: Quantity{resource.MustParse("10Mi")},
			},
		},
		BuildJob: BuildJobResourceConfig{
			Resources: Resources{
				MinRequestedCPU:    Quantity{resource.MustParse("20m")},
				MinRequestedMemory: Quantity{resource.MustParse("20Mi")},
			},
		},
	}

	k8sClient := fake.NewClientBuilder().Build()
	require.NoError(t, serverlessv1alpha2.AddToScheme(scheme.Scheme))
	resourceClient := serverlessResource.New(k8sClient, scheme.Scheme)

	statsCollector := &automock.StatsCollector{}
	statsCollector.On("UpdateReconcileStats", mock.Anything, mock.Anything).Return()

	testCases := map[string]struct {
		fn              serverlessv1alpha2.Function
		expectedCondMsg string
	}{
		"Function requests cpu are bigger than limits": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-fn",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Function: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("120m"),
								corev1.ResourceMemory: resource.MustParse("120Mi"),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("150m"),
								corev1.ResourceMemory: resource.MustParse("50Mi"),
							}}},
					},
				},
			},
			expectedCondMsg: "Function limits cpu(120m) should be higher than requests cpu(150m)",
		},
		"Function requests memory are bigger than limits": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-fn",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Function: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("120m"),
								corev1.ResourceMemory: resource.MustParse("120Mi"),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("50m"),
								corev1.ResourceMemory: resource.MustParse("150Mi"),
							}}},
					},
				},
			},
			expectedCondMsg: "Function limits memory(120Mi) should be higher than requests memory(150Mi)",
		},
		"Function requests cpu are smaller than minimum value": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-fn",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Function: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("120m"),
								corev1.ResourceMemory: resource.MustParse("120Mi"),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("5m"),
								corev1.ResourceMemory: resource.MustParse("50Mi"),
							}}},
					},
				},
			},
			expectedCondMsg: "Function request cpu(5m) should be higher than minimal value (10m)",
		},
		"Function requests memory are smaller than minimum value": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-fn",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Function: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("120m"),
								corev1.ResourceMemory: resource.MustParse("120Mi"),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("50m"),
								corev1.ResourceMemory: resource.MustParse("5Mi"),
							}}},
					},
				},
			},
			expectedCondMsg: "Function request memory(5Mi) should be higher than minimal value (10Mi)",
		},
		"Function limits cpu are smaller than minimum without requests": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-fn",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Function: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("2m"),
								corev1.ResourceMemory: resource.MustParse("120Mi"),
							},
						}},
					},
				},
			},
			expectedCondMsg: "Function limits cpu(2m) should be higher than minimal value (10m)",
		},
		"Function limits memory are smaller than minimum without requests": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-fn",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Function: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("20m"),
								corev1.ResourceMemory: resource.MustParse("2Mi"),
							},
						}},
					},
				},
			},
			expectedCondMsg: "Function limits memory(2Mi) should be higher than minimal value (10Mi)",
		},
		//Build validation
		"Build requests cpu are bigger than limits": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-fn",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("120m"),
								corev1.ResourceMemory: resource.MustParse("120Mi"),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("150m"),
								corev1.ResourceMemory: resource.MustParse("50Mi"),
							}}},
					},
				},
			},
			expectedCondMsg: "Build limits cpu(120m) should be higher than requests cpu(150m)",
		},
		"Build requests memory are bigger than limits": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-fn",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("120m"),
								corev1.ResourceMemory: resource.MustParse("120Mi"),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("50m"),
								corev1.ResourceMemory: resource.MustParse("150Mi"),
							}}},
					},
				},
			},
			expectedCondMsg: "Build limits memory(120Mi) should be higher than requests memory(150Mi)",
		},
		"Build requests cpu are smaller than minimum value": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-fn",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("120m"),
								corev1.ResourceMemory: resource.MustParse("120Mi"),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("5m"),
								corev1.ResourceMemory: resource.MustParse("50Mi"),
							}}},
					},
				},
			},
			expectedCondMsg: "Build request cpu(5m) should be higher than minimal value (20m)",
		},
		"Build requests memory are smaller than minimum value": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-fn",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("120m"),
								corev1.ResourceMemory: resource.MustParse("120Mi"),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("50m"),
								corev1.ResourceMemory: resource.MustParse("5Mi"),
							}}},
					},
				},
			},
			expectedCondMsg: "Build request memory(5Mi) should be higher than minimal value (20Mi)",
		},
		"Build limits cpu are smaller than minimum without requests": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-fn",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("2m"),
								corev1.ResourceMemory: resource.MustParse("120Mi"),
							},
						}},
					},
				},
			},
			expectedCondMsg: "Build limits cpu(2m) should be higher than minimal value (20m)",
		},
		"Build limits memory are smaller than minimum without requests": {
			fn: serverlessv1alpha2.Function{ObjectMeta: metav1.ObjectMeta{GenerateName: "test-fn"},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("30m"),
								corev1.ResourceMemory: resource.MustParse("2Mi"),
							},
						}},
					},
				},
			},
			expectedCondMsg: "Build limits memory(2Mi) should be higher than minimal value (20Mi)",
		},
		"Invalid env": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{GenerateName: "test-fn"},
				Spec: serverlessv1alpha2.FunctionSpec{
					Env: []corev1.EnvVar{
						{Name: "1ENV"},
						{Name: "2ENV"},
					},
				},
			},
			expectedCondMsg: "spec.env: 1ENV. Err: a valid environment variable name must consist of alphabetic characters, digits, '_', '-', or '.', and must not start with a digit (e.g. 'my.env-name',  or 'MY_ENV.NAME',  or 'MyEnvName1', regex used for validation is '[-._a-zA-Z][-._a-zA-Z0-9]*')",
		},
	}

	//WHEN
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, resourceClient.Create(ctx, &tc.fn))
			s := &systemState{instance: tc.fn}
			r := &reconciler{out: out{result: controllerruntime.Result{Requeue: true}},
				k8s: k8s{client: resourceClient, recorder: record.NewFakeRecorder(100), statsCollector: statsCollector},
				cfg: cfg{fn: FunctionConfig{ResourceConfig: minResourcesCfg}}}

			//WHEN
			nextFn, err := stateFnValidateFunction(ctx, r, s)
			require.NoError(t, err)
			_, err = nextFn(context.TODO(), r, s)

			//THEN
			require.NoError(t, err)
			updatedFn := serverlessv1alpha2.Function{}
			require.NoError(t, resourceClient.Get(ctx, ctrlclient.ObjectKey{Name: tc.fn.Name}, &updatedFn))
			cond := getCondition(updatedFn.Status.Conditions, serverlessv1alpha2.ConditionConfigurationReady)
			assert.Equal(t, serverlessv1alpha2.ConditionReasonFunctionSpec, cond.Reason)
			assert.Equal(t, corev1.ConditionFalse, cond.Status)
			assert.NotEmpty(t, tc.expectedCondMsg, "expected message shouldn't be empty")
			assert.Equal(t, tc.expectedCondMsg, cond.Message)
			assert.False(t, r.result.Requeue)

		})
	}

	t.Run("Valid function", func(t *testing.T) {
		//GIVEN
		fn := serverlessv1alpha2.Function{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-fn",
			},
			Spec: serverlessv1alpha2.FunctionSpec{
				ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
					Build: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
						Limits: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceCPU:    resource.MustParse("120m"),
							corev1.ResourceMemory: resource.MustParse("120Mi"),
						},
						Requests: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("100Mi"),
						}}},
				},
				Env: []corev1.EnvVar{
					{Name: "_CORRECT_ENV"},
					{Name: "ANOTHER_CORRECT_ENV"},
				},
				SecretMounts: []serverlessv1alpha2.SecretMount{
					{
						SecretName: "secret-name",
						MountPath:  "mount-path",
					},
				},
			},
		}
		r := &reconciler{out: out{result: controllerruntime.Result{Requeue: true}},
			k8s: k8s{client: resourceClient, recorder: record.NewFakeRecorder(100), statsCollector: statsCollector},
			cfg: cfg{fn: FunctionConfig{ResourceConfig: minResourcesCfg}}}

		//WHEN
		require.NoError(t, resourceClient.Create(ctx, &fn))
		s := &systemState{instance: fn}
		_, err := stateFnValidateFunction(ctx, r, s)

		//THEN
		require.NoError(t, err)
		updatedFn := serverlessv1alpha2.Function{}
		require.NoError(t, resourceClient.Get(ctx, ctrlclient.ObjectKey{Name: fn.Name}, &updatedFn))
		assert.True(t, r.result.Requeue)

	})
}
