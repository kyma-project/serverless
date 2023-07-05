package annotation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	testDeployCR = &appsv1.Deployment{
		TypeMeta: v1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-deploy",
			Namespace: "default",
			Annotations: map[string]string{
				"reconciler.kyma-project.io/managed-by-reconciler-disclaimer": "test message",
			},
		},
	}
)

func TestAddDoNotEditDisclaimer(t *testing.T) {
	t.Run("add disclaimer", func(t *testing.T) {
		obj := unstructured.Unstructured{}
		obj = AddDoNotEditDisclaimer(obj)

		require.Equal(t, message, obj.GetAnnotations()[annotation])
	})
}

func TestDeleteReconcilerDisclaimer(t *testing.T) {
	t.Run("remove annotation", func(t *testing.T) {
		scheme := runtime.NewScheme()
		appsv1.AddToScheme(scheme)

		client := fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(testDeployCR).
			Build()

		obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(testDeployCR)
		require.NoError(t, err)

		err = DeleteReconcilerDisclaimer(client, rest.Config{}, unstructured.Unstructured{Object: obj})
		require.NoError(t, err)

		var expected appsv1.Deployment
		err = client.Get(context.Background(),
			types.NamespacedName{
				Namespace: testDeployCR.Namespace,
				Name:      testDeployCR.Name,
			}, &expected)
		require.NoError(t, err)
		require.Len(t, expected.Annotations, 0)
	})
}
