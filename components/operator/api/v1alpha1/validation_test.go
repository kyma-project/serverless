package v1alpha1_test

import (
	"context"
	"testing"

	v1alpha1 "github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/components/operator/internal/testenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	enableInternal := true
	//GIVEN
	testCases := map[string]struct {
		serverless *v1alpha1.Serverless
	}{
		"No PersistenceVolume explicit config": {
			serverless: &v1alpha1.Serverless{
				ObjectMeta: fixMetadata,
				Spec: v1alpha1.ServerlessSpec{
					DockerRegistry: &v1alpha1.DockerRegistry{
						EnableInternal: &enableInternal,
					},
				},
			},
		},
		"Explicit PersistenceVolume config": {
			serverless: &v1alpha1.Serverless{
				ObjectMeta: fixMetadata,
				Spec: v1alpha1.ServerlessSpec{
					DockerRegistry: &v1alpha1.DockerRegistry{
						EnableInternal: &enableInternal,
						PersistenceVolume: &v1alpha1.PersistenceVolume{
							Size: resource.MustParse("2Gi"),
						},
					},
				},
			},
		},
		"Empty PersistenceVolume config": {
			serverless: &v1alpha1.Serverless{
				ObjectMeta: fixMetadata,
				Spec: v1alpha1.ServerlessSpec{
					DockerRegistry: &v1alpha1.DockerRegistry{
						EnableInternal:    &enableInternal,
						PersistenceVolume: &v1alpha1.PersistenceVolume{},
					},
				},
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			//WHEN
			err := k8sClient.Create(ctx, tc.serverless)
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
	enableInternal := false
	//GIVEN
	testCases := map[string]struct {
		serverless     *v1alpha1.Serverless
		expectedErrMsg string
		fieldPath      string
		expectedCause  metav1.CauseType
	}{
		"PersistenceVolume size config w/o enableInternal": {
			serverless: &v1alpha1.Serverless{
				ObjectMeta: fixMetadata,
				Spec: v1alpha1.ServerlessSpec{
					DockerRegistry: &v1alpha1.DockerRegistry{
						PersistenceVolume: &v1alpha1.PersistenceVolume{
							Size: resource.MustParse("2Gi"),
						},
					},
				},
			},
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
			expectedErrMsg: "Use dockerRegistry.persistenceVolume only in combination with dockerRegistry.enableInternal set to true",
			fieldPath:      "spec.dockerRegistry",
		},
		"PersistenceVolume size config and enableInternal=false": {
			serverless: &v1alpha1.Serverless{
				ObjectMeta: fixMetadata,
				Spec: v1alpha1.ServerlessSpec{
					DockerRegistry: &v1alpha1.DockerRegistry{
						EnableInternal: &enableInternal,
						PersistenceVolume: &v1alpha1.PersistenceVolume{
							Size: resource.MustParse("2Gi"),
						},
					},
				},
			},
			expectedCause:  metav1.CauseTypeFieldValueInvalid,
			expectedErrMsg: "Use dockerRegistry.persistenceVolume only in combination with dockerRegistry.enableInternal set to true",
			fieldPath:      "spec.dockerRegistry",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			//WHEN
			err := k8sClient.Create(ctx, tc.serverless)

			//THEN
			require.Error(t, err)
			errStatus, ok := err.(*k8serrors.StatusError)
			require.True(t, ok)
			causes := errStatus.Status().Details.Causes
			require.Len(t, causes, 1)
			cause := causes[0]
			assert.Equal(t, tc.expectedCause, cause.Type)
			assert.Equal(t, tc.fieldPath, cause.Field)
			assert.NotEmpty(t, tc.expectedErrMsg, "cause message: %s", cause.Message)
			//TODO: better will be Equal comparison
			assert.Contains(t, cause.Message, tc.expectedErrMsg)
		})
	}
}
