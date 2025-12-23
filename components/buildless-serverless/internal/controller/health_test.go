package controller

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

func TestHealthChecker_Checker(t *testing.T) {
	log := zap.NewNop().Sugar()

	t.Run("Metrics success", func(t *testing.T) {
		registerPrometheusGauges(1, 2)
		//GIVEN
		timeout := 10 * time.Second
		checker, _, _ := NewHealthChecker(timeout, log)

		//WHEN
		err := checker.Checker(nil)

		//THEN
		require.NoError(t, err)
	})

	t.Run("Metrics stuck", func(t *testing.T) {
		registerPrometheusGauges(1, 0)
		//GIVEN
		timeout := 10 * time.Second
		checker, _, _ := NewHealthChecker(timeout, log)

		//WHEN
		err := checker.Checker(nil)

		//THEN
		require.Contains(t, err.Error(), "reconcile loop is stuck based on metrics")
	})

	t.Run("Event success", func(t *testing.T) {
		registerPrometheusGauges(0, 0)
		//GIVEN
		timeout := 10 * time.Second
		checker, inCh, outCh := NewHealthChecker(timeout, log)

		//WHEN
		go func() {
			check := <-inCh
			require.Equal(t, check.Object.GetName(), HealthEvent)
			outCh <- true
		}()
		err := checker.Checker(nil)

		//THEN
		require.NoError(t, err)
	})

	t.Run("Event timeout", func(t *testing.T) {
		registerPrometheusGauges(0, 0)
		//GIVEN
		timeout := time.Second
		checker, inCh, _ := NewHealthChecker(timeout, log)

		//WHEN
		go func() {
			check := <-inCh
			require.Equal(t, check.Object.GetName(), HealthEvent)
		}()
		err := checker.Checker(nil)

		//THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "reconcile didn't send confirmation")

	})

	t.Run("Can't send check event", func(t *testing.T) {
		registerPrometheusGauges(0, 0)
		//GIVEN
		timeout := time.Second
		checker, _, _ := NewHealthChecker(timeout, log)

		//WHEN
		err := checker.Checker(nil)

		//THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "timeout when sending check event")
	})
}

func TestHealthName(t *testing.T) {
	//GIVEN
	//WHEN
	// This const is longer than 253 characters to avoid collisions with real k8s objects.
	require.Greater(t, len(HealthEvent), 253)
	//THEN
}

func registerPrometheusGauges(workqueueDepth int, succesfullReconciliations int) {
	metrics.Registry = prometheus.NewRegistry()

	workqueueGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "workqueue_depth",
		Help: "Current depth of workqueue",
	}, []string{"controller", "name"})
	workqueueGauge.WithLabelValues("function", "function").Set(float64(workqueueDepth))
	metrics.Registry.MustRegister(workqueueGauge)

	reconcileTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "controller_runtime_reconcile_total",
			Help: "Total number of reconciliations per controller",
		},
		[]string{"controller", "result"},
	)
	reconcileTotal.WithLabelValues("function", "error").Add(0)
	reconcileTotal.WithLabelValues("function", "requeue").Add(0)
	reconcileTotal.WithLabelValues("function", "requeueAfter").Add(0)
	reconcileTotal.WithLabelValues("function", "success").Add(float64(succesfullReconciliations))
	metrics.Registry.MustRegister(reconcileTotal)
}
