package state

import (
	"context"
	"testing"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/components/operator/internal/chart"
	"github.com/kyma-project/serverless/components/operator/internal/warning"
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
	t.Run("stop when serverless in ready", func(t *testing.T) {
		s := &systemState{
			warningBuilder: warning.NewBuilder(),
			instance:       *testInstalledServerless.DeepCopy(),
			chartConfig: &chart.Config{
				Cache: fixEmptyManifestCache(),
				CacheKey: types.NamespacedName{
					Name:      testInstalledServerless.GetName(),
					Namespace: testInstalledServerless.GetNamespace(),
				},
			},
		}
		r := &reconciler{
			log: zap.NewNop().Sugar(),
			k8s: k8s{
				client: fake.NewClientBuilder().Build(),
			},
		}

		// verify and return update condition state
		next, result, err := sFnVerifyResources(context.Background(), r, s)
		require.Nil(t, err)
		require.Nil(t, result)
		require.Nil(t, next)

		status := s.instance.Status
		require.Equal(t, v1alpha1.StateReady, status.State)
		require.Len(t, status.Conditions, 2)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeInstalled,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonInstallation,
			"",
		)
	})

	t.Run("set to ready and requeue", func(t *testing.T) {
		serverless := &v1alpha1.Serverless{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
		}
		s := &systemState{
			warningBuilder: warning.NewBuilder(),
			instance:       *serverless,
			chartConfig: &chart.Config{
				Cache: fixEmptyManifestCache(),
				CacheKey: types.NamespacedName{
					Name:      serverless.GetName(),
					Namespace: serverless.GetNamespace(),
				},
			},
		}
		r := &reconciler{
			log: zap.NewNop().Sugar(),
			k8s: k8s{
				client: fake.NewClientBuilder().Build(),
			},
		}

		// verify and return update condition state
		next, result, err := sFnVerifyResources(context.Background(), r, s)
		require.Nil(t, err)
		require.Equal(t, requeueResult, result)
		require.Nil(t, next)

		status := s.instance.Status
		require.Equal(t, v1alpha1.StateReady, status.State)
		require.Len(t, status.Conditions, 1)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeInstalled,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonInstalled,
			"Serverless installed",
		)
	})

	t.Run("warning", func(t *testing.T) {
		s := &systemState{
			warningBuilder: warning.NewBuilder().With("test warning"),
			instance:       *testInstalledServerless.DeepCopy(),
			chartConfig: &chart.Config{
				Cache: fixEmptyManifestCache(),
				CacheKey: types.NamespacedName{
					Name:      testInstalledServerless.GetName(),
					Namespace: testInstalledServerless.GetNamespace(),
				},
			},
		}
		r := &reconciler{
			log: zap.NewNop().Sugar(),
		}

		// verify and return update condition state
		next, result, err := sFnVerifyResources(context.Background(), r, s)
		require.Nil(t, err)
		require.Nil(t, result)
		require.Nil(t, next)

		status := s.instance.Status
		require.Equal(t, v1alpha1.StateWarning, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeInstalled,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonInstalled,
			s.warningBuilder.Build(),
		)
	})

	t.Run("verify error", func(t *testing.T) {
		s := &systemState{
			instance: *testInstalledServerless.DeepCopy(),
			chartConfig: &chart.Config{
				Cache: fixManifestCache("\t"),
				CacheKey: types.NamespacedName{
					Name:      testInstalledServerless.GetName(),
					Namespace: testInstalledServerless.GetNamespace(),
				},
			},
		}
		r := &reconciler{
			log: zap.NewNop().Sugar(),
		}

		// handle verify err and update condition with err
		next, result, err := sFnVerifyResources(context.Background(), r, s)
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

	t.Run("requeue when resources are not ready", func(t *testing.T) {
		client := fake.NewClientBuilder().WithObjects(testDeployCR).Build()
		s := &systemState{
			instance: *testInstalledServerless.DeepCopy(),
			chartConfig: &chart.Config{
				Cache: func() chart.ManifestCache {
					cache := chart.NewInMemoryManifestCache()
					_ = cache.Set(context.Background(), types.NamespacedName{
						Name:      testInstalledServerless.GetName(),
						Namespace: testInstalledServerless.GetNamespace(),
					}, chart.ServerlessSpecManifest{Manifest: testDeployManifest})
					return cache
				}(),
				CacheKey: types.NamespacedName{
					Name:      testInstalledServerless.GetName(),
					Namespace: testInstalledServerless.GetNamespace(),
				},
				Cluster: chart.Cluster{
					Client: client,
				},
			},
		}
		r := &reconciler{}

		// return requeue on verification failed
		next, result, err := sFnVerifyResources(context.Background(), r, s)

		_, expectedResult, _ := requeueAfter(requeueDuration)
		require.NoError(t, err)
		require.Equal(t, expectedResult, result)
		require.Nil(t, next)
	})
}
