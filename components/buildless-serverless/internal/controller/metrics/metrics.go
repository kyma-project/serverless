package metrics

import (
	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"time"
)

type functionStateReachTimeInfo struct {
	startTime           *time.Time
	generation          int64
	registeredCondition *serverlessv1alpha2.ConditionType
}

var (
	FunctionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "serverless_function_processed_total",
			Help: "Total number of functions processed (each function is counted only once, even if it is processed multiple times)",
		},
		[]string{"runtime", "source"},
	)
	ReconciliationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "serverless_function_reconciliations_total",
			Help: "Total number of reconciliations for functions (each reconciliation is counted, even if it is for the same function)",
		},
		[]string{"runtime", "source"},
	)
	ReconciliationTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "serverless_function_reconciliation_time_seconds",
			Help:    "Time taken for a single reconciliation of a function (including all single reconciliation)",
			Buckets: []float64{0.1, 0.3, 1, 3, 10, 30, 90, 300},
		},
		[]string{"runtime", "source"},
	)
	StateReachTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "serverless_function_state_reach_time_seconds",
			Help:    "Time taken for a function to reach a specific state (only generation changes are counted, not every reconciliation)",
			Buckets: []float64{0.1, 0.3, 1, 3, 10, 30, 90, 300},
		},
		[]string{"runtime", "source", "state"},
	)
	stateReachTimeInfo     = map[string]functionStateReachTimeInfo{}
	processedFunctionsUIDs = sets.Set[string]{}
)

func Register() {
	metrics.Registry.MustRegister(
		FunctionsTotal,
		ReconciliationsTotal,
		ReconciliationTime,
		StateReachTime,
	)
}

func runtimeName(f serverlessv1alpha2.Function) string {
	return string(f.Spec.Runtime)
}

func sourceType(f serverlessv1alpha2.Function) string {
	if f.HasGitSources() {
		return "git"
	}
	return "inline"
}

func PublishFunctionsTotal(f serverlessv1alpha2.Function) {
	uid := string(f.UID)
	if processedFunctionsUIDs.Has(uid) {
		return // Function already processed, no need to increment
	}
	processedFunctionsUIDs.Insert(uid)
	FunctionsTotal.WithLabelValues(runtimeName(f), sourceType(f)).Inc()
}

func PublishReconciliationsTotal(f serverlessv1alpha2.Function) {
	ReconciliationsTotal.WithLabelValues(runtimeName(f), sourceType(f)).Inc()
}

func PublishReconciliationTime(f serverlessv1alpha2.Function, start time.Time) {
	duration := time.Since(start).Seconds()
	ReconciliationTime.WithLabelValues(runtimeName(f), sourceType(f)).Observe(duration)
}

func StartForStateReachTime(f serverlessv1alpha2.Function) {
	uid := string(f.UID)
	fi, ok := stateReachTimeInfo[uid]
	if !ok {
		fi = functionStateReachTimeInfo{}
	}
	if fi.generation == f.GetGeneration() {
		return // Function already measured, no need to set start time
	}
	if fi.generation == 0 && f.Status.ObservedGeneration == f.GetGeneration() {
		return // Function already observed, no need to set start time
	}
	fi.startTime = ptr.To(time.Now())
	fi.generation = f.GetGeneration()
	fi.registeredCondition = nil
	stateReachTimeInfo[uid] = fi
}

// conditionIsGreaterThanRegistered returns true if cond is greater than registeredCond according to the order: nil < ConfigurationReady < Running < Ready
func conditionIsGreaterThanRegistered(cond serverlessv1alpha2.ConditionType, registeredCond *serverlessv1alpha2.ConditionType) bool {
	order := map[serverlessv1alpha2.ConditionType]int{
		serverlessv1alpha2.ConditionConfigurationReady: 1,
		serverlessv1alpha2.ConditionRunning:            2,
	}
	var regVal int
	if registeredCond == nil {
		regVal = 0
	} else {
		regVal = order[*registeredCond]
	}
	return order[cond] > regVal
}

func PublishStateReachTime(f serverlessv1alpha2.Function, toState serverlessv1alpha2.ConditionType) {
	uid := string(f.UID)
	fi, ok := stateReachTimeInfo[uid]
	if !ok || fi.startTime == nil {
		return // No start time recorded, nothing to publish
	}
	if !conditionIsGreaterThanRegistered(toState, fi.registeredCondition) {
		return // duplicated state publishing, nothing to publish
	}
	fi.registeredCondition = &toState
	stateReachTimeInfo[uid] = fi

	duration := time.Since(*fi.startTime).Seconds()
	StateReachTime.WithLabelValues(runtimeName(f), sourceType(f), string(toState)).Observe(duration)
}
