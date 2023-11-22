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

	k8sClient := fake.NewClientBuilder().Build()
	require.NoError(t, serverlessv1alpha2.AddToScheme(scheme.Scheme))
	resourceClient := serverlessResource.New(k8sClient, scheme.Scheme)

	statsCollector := &automock.StatsCollector{}
	statsCollector.On("UpdateReconcileStats", mock.Anything, mock.Anything).Return()

	r := &reconciler{out: out{result: controllerruntime.Result{}},
		k8s: k8s{client: resourceClient, recorder: record.NewFakeRecorder(100), statsCollector: statsCollector}}

	testCases := map[string]struct {
		fn serverlessv1alpha2.Function
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
		},
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
		},
	}

	//WHEN
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, resourceClient.Create(ctx, &tc.fn))
			s := &systemState{instance: tc.fn}
			nextFn, err := stateFnValidateFunction(ctx, r, s)
			require.NoError(t, err)
			_, err = nextFn(context.TODO(), r, s)
			//THEN
			require.NoError(t, err)
			updatedFn := serverlessv1alpha2.Function{}
			require.NoError(t, resourceClient.Get(ctx, ctrlclient.ObjectKey{Name: tc.fn.Name}, &updatedFn))
			reason := getConditionReason(updatedFn.Status.Conditions, serverlessv1alpha2.ConditionConfigurationReady)
			assert.Contains(t, reason, serverlessv1alpha2.ConditionReasonFunctionSpec)
			status := getConditionStatus(updatedFn.Status.Conditions, serverlessv1alpha2.ConditionConfigurationReady)
			assert.Equal(t, corev1.ConditionFalse, status)
			assert.False(t, r.result.Requeue)

		})
	}

}
