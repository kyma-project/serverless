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
	t.Run("check prerequisites", func(t *testing.T) {
		s := &systemState{
			instance: testInstalledServerless,
		}

		r := &reconciler{}

		stateFn := buildSFnPrerequisites(s)
		requireEqualFunc(t, sFnPrerequisites, stateFn)

		next, result, err := stateFn(nil, r, s)

		expectedNext := sFnUpdateProcessingTrueState(
			buildSFnApplyResources(s),
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonPrerequisitesMet,
			"All prerequisites met",
		)

		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})

	t.Run("check prerequisites error", func(t *testing.T) {
		s := &systemState{
			instance: testInstalledServerless,
		}
		s.instance.Spec.DockerRegistry.EnableInternal = pointer.Bool(true)

		r := &reconciler{
			k8s: k8s{
				client: fake.NewFakeClient(),
			},
		}

		stateFn := buildSFnPrerequisites(s)
		requireEqualFunc(t, sFnPrerequisites, stateFn)

		next, result, err := stateFn(nil, r, s)

		expectedNext := sFnUpdateErrorState(
			sFnRequeue(),
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonPrerequisitesErr,
			errors.New("test error"),
		)

		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})
}
