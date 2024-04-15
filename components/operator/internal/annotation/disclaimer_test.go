package annotation

import (
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
