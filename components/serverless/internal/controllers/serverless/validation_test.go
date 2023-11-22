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

// TODO: add separate use cases with memory and cpu
func TestValidation(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	minResourcesCfg := ResourceConfig{
		Function: FunctionResourceConfig{
			Resources: Resources{
				MinRequestedCPU:    Quantity{resource.MustParse("10m")},
				MinRequestedMemory: Quantity{resource.MustParse("10m")},
			},
		},
		BuildJob: BuildJobResourceConfig{
			Resources: Resources{
				MinRequestedCPU:    Quantity{resource.MustParse("20m")},
				MinRequestedMemory: Quantity{resource.MustParse("20m")},
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
								corev1.ResourceMemory: resource.MustParse("120m"),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("150m"),
								corev1.ResourceMemory: resource.MustParse("50m"),
							}}},
					},
				},
			},
			expectedCondMsg: "Request cpu cannot be bigger than limits cpu",
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
								corev1.ResourceMemory: resource.MustParse("120m"),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("50m"),
								corev1.ResourceMemory: resource.MustParse("150m"),
							}}},
					},
				},
			},
			expectedCondMsg: "Request memory cannot be bigger than limits memory",
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
								corev1.ResourceMemory: resource.MustParse("120m"),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("5m"),
								corev1.ResourceMemory: resource.MustParse("50m"),
							}}},
					},
				},
			},
			expectedCondMsg: "less than minimum cpu",
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
								corev1.ResourceMemory: resource.MustParse("120m"),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("50m"),
								corev1.ResourceMemory: resource.MustParse("5m"),
							}}},
					},
				},
			},
			expectedCondMsg: "Function request memory(5m) should be higher than minimal value",
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
								corev1.ResourceMemory: resource.MustParse("120m"),
							},
						}},
					},
				},
			},
			expectedCondMsg: "Limits cpu cannot be less than minimum cpu",
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
								corev1.ResourceMemory: resource.MustParse("2m"),
							},
						}},
					},
				},
			},
			expectedCondMsg: "Limits memory cannot be less than minimum memory",
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
								corev1.ResourceMemory: resource.MustParse("120m"),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("150m"),
								corev1.ResourceMemory: resource.MustParse("50m"),
							}}},
					},
				},
			},
			expectedCondMsg: "Request cpu cannot be bigger than limits cpu",
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
								corev1.ResourceMemory: resource.MustParse("120m"),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("50m"),
								corev1.ResourceMemory: resource.MustParse("150m"),
							}}},
					},
				},
			},
			expectedCondMsg: "Request memory cannot be bigger than limits memory",
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
								corev1.ResourceMemory: resource.MustParse("120m"),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("5m"),
								corev1.ResourceMemory: resource.MustParse("50m"),
							}}},
					},
				},
			},
			expectedCondMsg: "less than minimum cpu",
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
								corev1.ResourceMemory: resource.MustParse("120m"),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("50m"),
								corev1.ResourceMemory: resource.MustParse("5m"),
							}}},
					},
				},
			},
			expectedCondMsg: "less than minimum memory",
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
								corev1.ResourceMemory: resource.MustParse("120m"),
							},
						}},
					},
				},
			},
			expectedCondMsg: "Limits cpu cannot be less than minimum cpu",
		},
		"Build limits memory are smaller than minimum without requests": {
			fn: serverlessv1alpha2.Function{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-fn",
				},
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
						Build: &serverlessv1alpha2.ResourceRequirements{Resources: &corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("30m"),
								corev1.ResourceMemory: resource.MustParse("2m"),
							},
						}},
					},
				},
			},
			expectedCondMsg: "Limits memory cannot be less than minimum memory",
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
			assert.Equal(t, cond.Reason, serverlessv1alpha2.ConditionReasonFunctionSpec)
			assert.Equal(t, corev1.ConditionFalse, cond.Status)
			assert.NotEmpty(t, tc.expectedCondMsg, "expected message shouldn't be empty")
			assert.Contains(t, cond.Message, tc.expectedCondMsg)
			assert.False(t, r.result.Requeue)

		})
	}

	t.Run("Valid function resources", func(t *testing.T) {
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
							corev1.ResourceMemory: resource.MustParse("120m"),
						},
						Requests: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("100m"),
						}}},
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
