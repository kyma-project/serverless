package state

import (
	"context"
	"testing"

	"github.com/kyma-project/manager-toolkit/installation/chart"
	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/components/operator/internal/flags"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	envName = "IMAGE_FUNCTION_CONTROLLER"
)

func Test_buildSFnApplyResources(t *testing.T) {
	t.Run("switch state and add condition when condition is missing", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{},
			chartConfig: &chart.Config{
				Cache: fixEmptyManifestCache(),
				CacheKey: types.NamespacedName{
					Name:      testInstalledServerless.GetName(),
					Namespace: testInstalledServerless.GetNamespace(),
				},
				Release: chart.Release{
					Name:      testInstalledServerless.GetName(),
					Namespace: testInstalledServerless.GetNamespace(),
				},
			},
			flagsBuilder: flags.NewBuilder(),
		}

		r := &reconciler{
			k8s: k8s{
				client: fake.NewClientBuilder().Build(),
			},
		}

		next, result, err := sFnApplyResources(context.Background(), r, s)
		require.Nil(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnVerifyResources, next)

		expectedFlags := map[string]interface{}{
			"global": map[string]interface{}{
				"commonLabels": map[string]interface{}{
					"app.kubernetes.io/managed-by": "serverless-operator",
				},
			},
		}

		flags, err := s.flagsBuilder.Build()
		require.NoError(t, err)

		require.Equal(t, expectedFlags, flags)

		status := s.instance.Status
		require.Equal(t, v1alpha1.StateProcessing, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeInstalled,
			metav1.ConditionUnknown,
			v1alpha1.ConditionReasonInstallation,
			"Installing for configuration",
		)
	})

	t.Run("apply resources", func(t *testing.T) {
		s := &systemState{
			instance: *testInstalledServerless.DeepCopy(),
			chartConfig: &chart.Config{
				Cache: fixEmptyManifestCache(),
				CacheKey: types.NamespacedName{
					Name:      testInstalledServerless.GetName(),
					Namespace: testInstalledServerless.GetNamespace(),
				},
				Release: chart.Release{
					Name:      testInstalledServerless.GetName(),
					Namespace: testInstalledServerless.GetNamespace(),
				},
			},
			flagsBuilder: flags.NewBuilder(),
		}
		r := &reconciler{}

		// run installation process and return verificating state
		next, result, err := sFnApplyResources(context.Background(), r, s)
		require.Nil(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnVerifyResources, next)
	})

	t.Run("install chart error", func(t *testing.T) {
		s := &systemState{
			instance: *testInstalledServerless.DeepCopy(),
			chartConfig: &chart.Config{
				Cache: fixManifestCache("\t"),
				CacheKey: types.NamespacedName{
					Name:      testInstalledServerless.GetName(),
					Namespace: testInstalledServerless.GetNamespace(),
				},
			},
			flagsBuilder: flags.NewBuilder(),
		}
		r := &reconciler{
			log: zap.NewNop().Sugar(),
		}

		// handle error and return update condition state
		next, result, err := sFnApplyResources(context.Background(), r, s)
		require.EqualError(t, err, "could not parse chart manifest: yaml: found character that cannot start any token")
		require.Nil(t, result)
		require.Nil(t, next)

		status := s.instance.Status
		require.Equal(t, v1alpha1.StateError, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeInstalled,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonInstallationErr,
			"could not parse chart manifest: yaml: found character that cannot start any token",
		)
	})
}

func TestUpdateImageIfOverride(t *testing.T) {
	t.Run("Override image", func(t *testing.T) {
		t.Setenv(envName, "newImage")
		expectedFlags := map[string]interface{}{
			"global": map[string]interface{}{
				"images": map[string]interface{}{
					"function_buildful_controller": "newImage",
				},
			},
		}

		fb := flags.NewBuilder()

		updateImageIfOverride(envName, fb.WithImageFunctionBuildfulController)
		flags, err := fb.Build()
		require.NoError(t, err)
		require.Equal(t, expectedFlags, flags)
	})
	t.Run("Don't override image when empty env", func(t *testing.T) {
		expectedFlags := map[string]interface{}{}

		fb := flags.NewBuilder()

		updateImageIfOverride(envName, fb.WithImageFunctionBuildfulController)
		flags, err := fb.Build()
		require.NoError(t, err)
		require.Equal(t, expectedFlags, flags)
	})
}
