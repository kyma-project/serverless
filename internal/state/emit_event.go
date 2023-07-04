package state

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func emitEvent(m *reconciler, s *systemState) {
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
	s.saveSnapshot()
}

func eventType(condition metav1.Condition) string {
	eventType := "Normal"
	if condition.Status == metav1.ConditionFalse {
		eventType = "Warning"
	}
	return eventType
}
