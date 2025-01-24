package resources

import (
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
)

func TestNewService(t *testing.T) {
	t.Run("create proper service", func(t *testing.T) {
		f := &serverlessv1alpha2.Function{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-function-name",
				Namespace: "test-function-namespace",
				UID:       "test-uid",
			},
			Spec:   serverlessv1alpha2.FunctionSpec{},
			Status: serverlessv1alpha2.FunctionStatus{},
		}
		expectedSvc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-function-name",
				Namespace: "test-function-namespace",
				Labels: map[string]string{
					"serverless.kyma-project.io/function-name": "test-function-name",
					"serverless.kyma-project.io/managed-by":    "function-controller",
					"serverless.kyma-project.io/uuid":          "test-uid",
				},
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{{
					Name:       "http",
					TargetPort: intstr.FromInt32(8080),
					Port:       80,
					Protocol:   corev1.ProtocolTCP,
				}},
				Selector: map[string]string{
					"serverless.kyma-project.io/function-name": "test-function-name",
					"serverless.kyma-project.io/managed-by":    "function-controller",
					"serverless.kyma-project.io/resource":      "deployment",
					"serverless.kyma-project.io/uuid":          "test-uid",
				},
			},
		}

		r := NewService(f)

		require.NotNil(t, r)
		s := r.Service
		require.NotNil(t, s)
		require.IsType(t, &corev1.Service{}, s)
		require.Equal(t, expectedSvc, s)
	})
}
