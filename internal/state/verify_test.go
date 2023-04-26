package state

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	testDeployCR = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deploy",
			Namespace: "default",
		},
		Status: appsv1.DeploymentStatus{
			Conditions: []appsv1.DeploymentCondition{
				{
					Type:   appsv1.DeploymentAvailable,
					Status: corev1.ConditionUnknown,
				},
			},
		},
	}
)

const (
	testDeployManifest = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deploy
  namespace: default
`
)

func Test_sFnVerifyResources(t *testing.T) {
	t.Run("ready", func(t *testing.T) {
		s := &systemState{
			instance: testInstalledServerless,
			chartConfig: &chart.Config{
				Cache: testEmptyManifestCache(),
				Release: chart.Release{
					Name:      testInstalledServerless.GetName(),
					Namespace: testInstalledServerless.GetNamespace(),
				},
			},
		}

		r := &reconciler{
			log: zap.NewNop().Sugar(),
		}

		// build stateFn
		stateFn := buildSFnVerifyResources()

		// verify and return update condition state
		next, result, err := stateFn(context.Background(), r, s)

		expectedNext := sFnUpdateReadyState(
			sFnStop(),
			v1alpha1.ConditionTypeInstalled,
			v1alpha1.ConditionReasonInstalled,
			"Serverless installed",
		)

		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})

	t.Run("verify error", func(t *testing.T) {
		s := &systemState{
			instance: testInstalledServerless,
			chartConfig: &chart.Config{
				Cache:   testEmptyManifestCache(),
				Release: chart.Release{},
			},
		}

		r := &reconciler{
			log: zap.NewNop().Sugar(),
		}

		// build stateFn
		stateFn := buildSFnVerifyResources()

		// handle verify err and update condition with err
		next, result, err := stateFn(context.Background(), r, s)

		expectedNext := sFnUpdateErrorState(
			sFnRequeue(),
			v1alpha1.ConditionTypeInstalled,
			v1alpha1.ConditionReasonInstallationErr,
			errors.New("test err"),
		)

		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})

	t.Run("requeue when resources are not ready", func(t *testing.T) {
		client := fake.NewFakeClient(testDeployCR)

		s := &systemState{
			instance: testInstalledServerless,
			chartConfig: &chart.Config{
				Cache: func() *chart.ManifestCache {
					cache := chart.NewManifestCache()
					cache.Set(types.NamespacedName{
						Name:      testInstalledServerless.GetName(),
						Namespace: testInstalledServerless.GetNamespace(),
					}, nil, testDeployManifest)
					return cache
				}(),
				Release: chart.Release{
					Name:      testInstalledServerless.GetName(),
					Namespace: testInstalledServerless.GetNamespace(),
				},
				Cluster: chart.Cluster{
					Client: client,
				},
			},
		}

		r := &reconciler{}

		// build stateFn
		stateFn := buildSFnVerifyResources()

		// return requeue on verification failed
		next, result, err := stateFn(context.Background(), r, s)

		expectedNext, expectedResult, expectedErr := requeueAfter(requeueDuration)

		requireEqualFunc(t, expectedNext, next)
		require.Equal(t, expectedResult, result)
		require.Equal(t, expectedErr, err)
	})
}
