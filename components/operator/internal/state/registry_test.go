package state

import (
	"context"
	"testing"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/components/operator/internal/chart"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_sFnRegistryConfiguration(t *testing.T) {
	t.Run("internal registry and update", func(t *testing.T) {
		s := &systemState{
			instance:       v1alpha1.DockerRegistry{},
			statusSnapshot: v1alpha1.DockerRegistryStatus{},
			flagsBuilder:   chart.NewFlagsBuilder(),
		}
		r := &reconciler{
			k8s: k8s{client: fake.NewClientBuilder().Build()},
			log: zap.NewNop().Sugar(),
		}
		expectedFlags := map[string]interface{}{
			"dockerRegistry": map[string]interface{}{
				"enableInternal": true,
			},
			"global": map[string]interface{}{
				"registryNodePort": int64(32_137),
			},
		}

		next, result, err := sFnRegistryConfiguration(context.Background(), r, s)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnControllerConfiguration, next)

		require.EqualValues(t, expectedFlags, s.flagsBuilder.Build())
		require.Equal(t, v1alpha1.StateProcessing, s.instance.Status.State)
	})
}
