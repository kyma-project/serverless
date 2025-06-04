package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	ComponentVersion = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "serverless_version_info",
			Help: "Static metric with Serverless version info",
		},
		[]string{"version"},
	)
	ResourceProcessedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "serverless_resources_processed_total",
			Help: "Total number of resources processed for the first time, partitioned by user runtime",
		},
		[]string{"runtime"},
	)
)

func Register() {
	metrics.Registry.MustRegister(
		ComponentVersion,
		ResourceProcessedTotal,
	)
}
