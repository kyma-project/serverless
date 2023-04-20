package state

import (
	"errors"
	"testing"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_sFnPrerequisites(t *testing.T) {
	t.Run("update condition", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{},
		}

		r := &reconciler{}

		stateFn := sFnPrerequisites()
		next, result, err := stateFn(nil, r, s)

		expectedNext := sFnUpdateProcessingState(
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonConfigurationCheck,
			"Checking configuration",
		)

		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})
	t.Run("check prerequisites error", func(t *testing.T) {
		s := &systemState{
			instance: *testInstalledServerless.DeepCopy(),
		}
		s.instance.Spec.DockerRegistry.EnableInternal = pointer.Bool(true)

		r := &reconciler{
			k8s: k8s{
				client: fake.NewFakeClient(),
			},
		}

		stateFn := sFnPrerequisites()
		next, result, err := stateFn(nil, r, s)

		expectedNext := sFnUpdateErrorState(
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonConfigurationErr,
			errors.New("test error"),
		)

		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})
	t.Run("check prerequisites and return next state", func(t *testing.T) {
		s := &systemState{
			instance: testInstalledServerless,
		}

		r := &reconciler{}

		stateFn := sFnPrerequisites()
		next, result, err := stateFn(nil, r, s)

		expectedNext := sFnOptionalDependencies()

		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})
}
