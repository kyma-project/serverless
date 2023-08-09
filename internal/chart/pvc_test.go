package chart

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestPVC(t *testing.T) {
	testCases := map[string]struct {
		rawPVCToInstall *corev1.PersistentVolumeClaim
		clusterPVC      []client.Object
		expectedPVC     *corev1.PersistentVolumeClaim
	}{
		"pvc not exists in cluster": {
			rawPVCToInstall: fixPVC(20),
			expectedPVC:     fixPVC(20),
		},
		"pvc exists with the same size": {
			rawPVCToInstall: fixPVC(20),
			clusterPVC:      []client.Object{fixPVC(20)},
			expectedPVC:     fixPVC(20),
		},
		"pvc exists with bigger size": {
			rawPVCToInstall: fixPVC(20),
			clusterPVC:      []client.Object{fixPVC(30)},
			expectedPVC:     fixPVC(30),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			//GIVEN
			out, err := runtime.DefaultUnstructuredConverter.ToUnstructured(testCase.rawPVCToInstall)
			require.NoError(t, err)
			obj := unstructured.Unstructured{Object: out}

			c := fake.NewClientBuilder().WithObjects(testCase.clusterPVC...).Build()

			//WHEN
			finalObj, err := AdjustToClusterSize(context.TODO(), c, obj)

			//THEN
			require.NoError(t, err)

			expected, err := runtime.DefaultUnstructuredConverter.ToUnstructured(testCase.expectedPVC)

			require.NoError(t, err)
			require.EqualValues(t, expected, finalObj.Object)
		})
	}
}

func fixPVC(size int) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "registry-rawPVCToInstall",
			Namespace: "kyma-system",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(fmt.Sprintf("%dGi", size)),
				},
			},
		},
	}
}
