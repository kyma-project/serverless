package controller

import (
	"errors"
	"net/http"
	"time"

	io_prometheus_client "github.com/prometheus/client_model/go"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

const HealthEvent = "HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT_HEALTH_EVENT"

type ReconciliationMetrics struct {
	Errors       int64
	Requeue      int64
	RequeueAfter int64
	Success      int64
}

func (r ReconciliationMetrics) Total() int64 {
	return r.Errors + r.Requeue + r.RequeueAfter + r.Success
}

type HealthChecker struct {
	checkCh         chan event.GenericEvent
	healthCh        chan bool
	timeout         time.Duration
	log             *zap.SugaredLogger
	previousMetrics ReconciliationMetrics
}

func getHealthChannels() (chan event.GenericEvent, chan bool) {
	checkCh := make(chan event.GenericEvent)
	returnCh := make(chan bool)
	return checkCh, returnCh
}

func NewHealthChecker(timeout time.Duration, logger *zap.SugaredLogger) (HealthChecker, chan event.GenericEvent, chan bool) {
	checkCh, returnCh := getHealthChannels()
	emptyMetric := ReconciliationMetrics{}
	return HealthChecker{checkCh: checkCh, healthCh: returnCh, timeout: timeout, log: logger, previousMetrics: emptyMetric}, checkCh, returnCh
}

func findMetric(metrics []*io_prometheus_client.MetricFamily, name string) *io_prometheus_client.MetricFamily {
	for _, metric := range metrics {
		if metric.GetName() == name {
			return metric
		}
	}
	return nil
}

func parseReconciliationMetrics(metric *io_prometheus_client.MetricFamily) ReconciliationMetrics {
	reconcilations := ReconciliationMetrics{}
	for _, m := range metric.Metric {
		for _, label := range m.Label {
			if label.GetName() == "result" {
				switch label.GetValue() {
				case "error":
					reconcilations.Errors = int64(m.GetCounter().GetValue())
				case "requeue":
					reconcilations.Requeue = int64(m.GetCounter().GetValue())
				case "requeueAfter":
					reconcilations.RequeueAfter = int64(m.GetCounter().GetValue())
				case "success":
					reconcilations.Success = int64(m.GetCounter().GetValue())
				}
			}
		}
	}

	return reconcilations
}

func (h *HealthChecker) Checker(req *http.Request) error {
	h.log.Debug("Liveness handler triggered")
	// check in metrics if the module was doing something in a last few seconds
	metricsMissing := false

	allMetrics, err := metrics.Registry.Gather()
	if err != nil {
		h.log.Errorf("can't gather metrics: %v", err)
		metricsMissing = true
	}

	workqueueDepthMetric := findMetric(allMetrics, "workqueue_depth")
	if workqueueDepthMetric == nil {
		h.log.Error("can't find workqueue_depth metric")
		metricsMissing = true
	}

	totalReconcilesMetric := findMetric(allMetrics, "controller_runtime_reconcile_total")
	if totalReconcilesMetric == nil {
		h.log.Error("can't find controller_runtime_reconcile_total metric")
		metricsMissing = true
	}

	// in case metrics are missing skip directly to event-based check
	if !metricsMissing {
		totalReconciles := parseReconciliationMetrics(totalReconcilesMetric)
		workqueueDepth := int64(workqueueDepthMetric.Metric[0].Gauge.GetValue())

		// if the queue is not empty, check if the number of reconciliations has increased; update the state at the end
		defer func() {
			h.previousMetrics = totalReconciles
		}()
		if workqueueDepth > 0 {
			h.log.Debugf("workqueue not empty, depth %d, total reconciled prev -> now: %d -> %d", workqueueDepth, h.previousMetrics.Total(), totalReconciles.Total())
			if totalReconciles.Total() <= h.previousMetrics.Total() {
				h.log.Error("reconcile loop is stuck based on metrics")
				return errors.New("reconcile loop is stuck based on metrics")
			} else {
				h.log.Info("reconcile loop is healthy based on metrics")
				return nil
			}
		}
	}

	// fallback to sending an empty event to check if the module is alive when the queue is not empty
	// we don't want to use events by default in case the event queue is full, and the readiness check fails by timeout
	checkEvent := event.GenericEvent{
		Object: &corev1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name: HealthEvent,
			},
		},
	}
	select {
	case h.checkCh <- checkEvent:
	case <-time.After(h.timeout):
		return errors.New("timeout when sending check event")
	}

	h.log.Debug("check event sent to reconcile loop")
	select {
	case <-h.healthCh:
		h.log.Debug("reconcile loop is healthy")
		return nil
	case <-time.After(h.timeout):
		h.log.Debug("reconcile timeout")
		return errors.New("reconcile didn't send confirmation")
	}
}

func IsHealthCheckRequest(req ctrl.Request) bool { return req.Name == HealthEvent }
