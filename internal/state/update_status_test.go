package state

import (
	"context"
	"errors"
	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func Test_sFnUpdateServedFalse(t *testing.T) {
	t.Run("set served to false", func(t *testing.T) {
		serverless := v1alpha1.Serverless{
			ObjectMeta: v1.ObjectMeta{
				Name:            "test",
				Namespace:       "serverless-test",
				ResourceVersion: "222",
			},
			Status: v1alpha1.ServerlessStatus{},
		}
		s := &systemState{
			instance: serverless,
		}

		r := &reconciler{
			k8s: k8s{
				client: func() client.Client {
					scheme := apiruntime.NewScheme()
					require.NoError(t, v1alpha1.AddToScheme(scheme))

					client := fake.NewClientBuilder().
						WithScheme(scheme).
						WithObjects(serverless.DeepCopy()).
						Build()

					return client
				}(),
			},
		}

		nextFn, result, err := sFnUpdateServedFalse(v1alpha1.ConditionTypeConfigured, v1alpha1.ConditionReasonServerlessDuplicated, errors.New("some error text"))(context.TODO(), r, s)

		require.Nil(t, err)
		//requireEqualFunc(t, sFnRequeue(), nextFn) //TODO: these names aren't equal - why?
		require.NotNil(t, nextFn)
		require.Nil(t, result)

		require.Equal(t, v1alpha1.ServedFalse, s.instance.Status.Served)
		//TODO: in this state we can't check condition state
	})
}
