package state

import (
	"context"
	"testing"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/fsm"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/metrics"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_sFnMetricsStart(t *testing.T) {
	t.Run("metrics are updated", func(t *testing.T) {
		// Arrange
		// machine with our function
		f := serverlessv1alpha2.Function{
			ObjectMeta: metav1.ObjectMeta{
				Name: "blissful-jennings"},
			Spec: serverlessv1alpha2.FunctionSpec{
				Runtime:              "modest-spence",
				RuntimeImageOverride: "exciting-babbage",
				Source: serverlessv1alpha2.Source{
					Inline: &serverlessv1alpha2.InlineSource{
						Source: "dazzling-matsumoto"}}},
			Status: serverlessv1alpha2.FunctionStatus{}}
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Function: f}}

		// Act
		next, result, err := sFnMetricsStart(context.Background(), &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// without stopping processing
		require.Nil(t, result)
		// with expected next state
		require.NotNil(t, next)
		requireEqualFunc(t, sFnCleanupLegacyLeftovers, next)
		// metrics are set
		require.Equal(t, float64(1), testutil.ToFloat64(metrics.ReconciliationsTotal))
		require.Equal(t, float64(1), testutil.ToFloat64(metrics.FunctionsTotal))
	})
}
