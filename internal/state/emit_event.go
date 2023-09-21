package state

import (
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	warningMessagePrefix = "Warning"
)

func emitEvent(m *reconciler, s *systemState) {
	// compare if any condition change
	for _, condition := range s.instance.Status.Conditions {
		// check if condition exists in memento status
		memorizedCondition := meta.FindStatusCondition(s.statusSnapshot.Conditions, condition.Type)
		// ignore unchanged conditions
		if memorizedCondition != nil &&
			memorizedCondition.Status == condition.Status &&
			memorizedCondition.Reason == condition.Reason &&
			memorizedCondition.Message == condition.Message {
			continue
		}
		m.Event(
			&s.instance,
			eventType(condition, condition.Message),
			condition.Reason,
			condition.Message,
		)
	}
}

func eventType(condition metav1.Condition, message string) string {
	eventType := "Normal"
	if condition.Status == metav1.ConditionFalse || strings.HasPrefix(message, warningMessagePrefix) {
		eventType = "Warning"
	}
	return eventType
}
