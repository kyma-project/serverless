package v1alpha2_test

import (
	"context"
	"github.com/kyma-project/kyma/components/function-controller/internal/testenv"
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_XKubernetesValidations_Valid(t *testing.T) {
	fixMetadata := metav1.ObjectMeta{
		GenerateName: "test",
		Namespace:    "test",
	}
	ctx := context.TODO()
	k8sClient, testEnv := testenv.Start(t)
	defer testenv.Stop(t, testEnv)

	testNs := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
	}
	err := k8sClient.Create(ctx, &testNs)
	require.NoError(t, err)
	//GIVEN
	testCases := map[string]struct {
		fn *serverlessv1alpha2.Function
	}{
		"Profile set only for function": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{Function: &serverlessv1alpha2.ResourceRequirements{
						Profile: "Test",
					}},
				},
			},
		},
		"Profile set only for buildJob": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{Build: &serverlessv1alpha2.ResourceRequirements{
						Profile: "Test",
					}},
				},
			},
		},
		"Resource set only for buildJob": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{Build: &serverlessv1alpha2.ResourceRequirements{
						Profile: "Test",
					}},
				},
			},
		},
		"Resource set only for function": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{Function: &serverlessv1alpha2.ResourceRequirements{
						Profile: "Test",
					}},
				},
			},
		},
		"labels not use restricted domain": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Labels: map[string]string{
						"my.label.com": "labelValue",
					},
				},
			},
		},
		"similar label not use restricted domain": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Labels: map[string]string{
						"serverless.kyma-project.label.com": "labelValue",
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			//WHEN
			err := k8sClient.Create(ctx, tc.fn)
			//THEN
			require.NoError(t, err)
		})
	}
}
func Test_XKubernetesValidations_Invalid(t *testing.T) {
	fixMetadata := metav1.ObjectMeta{
		GenerateName: "test",
		Namespace:    "test",
	}
	ctx := context.TODO()
	k8sClient, testEnv := testenv.Start(t)
	defer testenv.Stop(t, testEnv)
	testNs := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
	}
	err := k8sClient.Create(ctx, &testNs)
	require.NoError(t, err)

	//GIVEN
	testCases := map[string]struct {
		fn             *serverlessv1alpha2.Function
		expectedErrMsg string
		fieldPath      string
	}{
		"Resource and Profiles used together in function": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{Function: &serverlessv1alpha2.ResourceRequirements{
						Profile:   "Test",
						Resources: &corev1.ResourceRequirements{},
					}},
				},
			},
			expectedErrMsg: "Use profile or resources",
			fieldPath:      "spec.resourceConfiguration.function",
		},

		"Resource and Profiles used together in buildJob": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{Build: &serverlessv1alpha2.ResourceRequirements{
						Profile:   "Test",
						Resources: &corev1.ResourceRequirements{},
					}},
				},
			},
			expectedErrMsg: "Use profile or resources",
			fieldPath:      "spec.resourceConfiguration.build",
		},
		"labels use exact restricted domain": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Labels: map[string]string{
						"serverless.kyma-project.io": "labelValue",
					},
				},
			},
			fieldPath:      "spec.labels",
			expectedErrMsg: "Labels has key starting with ",
		},
		"labels use restricted domain": {
			fn: &serverlessv1alpha2.Function{
				ObjectMeta: fixMetadata,
				Spec: serverlessv1alpha2.FunctionSpec{
					Labels: map[string]string{
						"serverless.kyma-project.io.com": "labelValue",
					},
				},
			},
			fieldPath:      "spec.labels",
			expectedErrMsg: "Labels has key starting with ",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			//WHEN
			err := k8sClient.Create(ctx, tc.fn)
			//THEN
			require.Error(t, err)
			errStatus, ok := err.(*k8serrors.StatusError)
			require.True(t, ok)
			causes := errStatus.Status().Details.Causes
			require.Len(t, causes, 1)
			cause := causes[0]
			assert.Equal(t, metav1.CauseTypeFieldValueInvalid, cause.Type)
			assert.Equal(t, tc.fieldPath, cause.Field)
			assert.Contains(t, cause.Message, tc.expectedErrMsg)
		})
	}
}
