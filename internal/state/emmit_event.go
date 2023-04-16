package state

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnEmitStrictEvent(next stateFn, result *ctrl.Result, err error, eventType, eventReason, eventMessage string) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		r.Event(
			&s.instance,
			eventType,
			eventReason,
			eventMessage,
		)

		return next, result, err
	}
}

func sFnEmitEventfunc(next stateFn, result *ctrl.Result, err error) stateFn {
	return func(_ context.Context, m *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		// compare if any condition change
		for _, condition := range s.instance.Status.Conditions {
			// check if condition exists in memento status
			memorizedCondition := meta.FindStatusCondition(s.snapshot.Conditions, condition.Type)
			// ignore unchanged conditions
			if memorizedCondition != nil &&
				memorizedCondition.Status == condition.Status &&
				memorizedCondition.Reason == condition.Reason &&
				memorizedCondition.Message == condition.Message {
				continue
			}
			m.Event(
				&s.instance,
				eventType(condition),
				condition.Reason,
				condition.Message,
			)
		}

		// take a snapshot to not repeat lastly emitted events
		return sFnTakeSnapshot(next, result, err)
	}
}

func eventType(condition metav1.Condition) string {
	eventType := "Normal"
	if condition.Status == metav1.ConditionFalse {
		eventType = "Warning"
	}
	return eventType
}
