package state

import (
	"context"
	"testing"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_sFnServedFilter(t *testing.T) {
	t.Run("re-processing when served is false", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{
				Status: v1alpha1.ServerlessStatus{
					Served: v1alpha1.ServedFalse,
				},
			},
		}

		r := &reconciler{
			k8s: k8s{
				client: fixClient(t),
			},
		}

		nextFn, result, err := sFnServedFilter(context.TODO(), r, s)
		require.Nil(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnAddFinalizer, nextFn)
	})

	t.Run("do next step when served is true", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{
				Status: v1alpha1.ServerlessStatus{
					Served: v1alpha1.ServedTrue,
				},
			},
		}

		nextFn, result, err := sFnServedFilter(context.TODO(), nil, s)
		require.Nil(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnAddFinalizer, nextFn)
	})

	t.Run("set served value from nil to true when there is no served serverless on cluster", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{
				Status: v1alpha1.ServerlessStatus{},
			},
		}

		r := &reconciler{
			k8s: k8s{
				client: fixClient(t,
					fixServedServerless("test-1", "default", ""),
					fixServedServerless("test-2", "serverless-test", ""),
					fixServedServerless("test-3", "serverless-test-2", ""),
					fixServedServerless("test-4", "default", ""),
				),
			},
		}

		nextFn, result, err := sFnServedFilter(context.TODO(), r, s)
		require.Nil(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnAddFinalizer, nextFn)
		require.Equal(t, v1alpha1.ServedTrue, s.instance.Status.Served)
	})

	t.Run("set served value from nil to false and set condition to error when there is at lease one served serverless on cluster", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{
				Status: v1alpha1.ServerlessStatus{},
			},
		}

		r := &reconciler{
			k8s: k8s{
				client: fixClient(t,
					fixServedServerless("test-1", "default", v1alpha1.ServedFalse),
					fixServedServerless("test-2", "serverless-test", v1alpha1.ServedTrue),
					fixServedServerless("test-3", "serverless-test-2", ""),
					fixServedServerless("test-4", "default", v1alpha1.ServedFalse),
				),
			},
		}

		nextFn, result, err := sFnServedFilter(context.TODO(), r, s)

		expectedErrorMessage := "only one instance of Serverless is allowed (current served instance: serverless-test/test-2) - this Serverless CR is redundant - remove it to fix the problem"
		require.EqualError(t, err, expectedErrorMessage)
		require.Nil(t, result)
		require.Nil(t, nextFn)
		require.Equal(t, v1alpha1.ServedFalse, s.instance.Status.Served)

		status := s.instance.Status
		require.Equal(t, v1alpha1.StateWarning, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeConfigured,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonServerlessDuplicated,
			expectedErrorMessage,
		)
	})
}

func fixClient(t *testing.T, initObjs ...client.Object) client.Client {
	scheme := apiruntime.NewScheme()
	require.NoError(t, v1alpha1.AddToScheme(scheme))

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(initObjs...).Build()
}

func fixServedServerless(name, namespace string, served v1alpha1.Served) *v1alpha1.Serverless {
	return &v1alpha1.Serverless{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Status: v1alpha1.ServerlessStatus{
			Served: served,
		},
	}
}
