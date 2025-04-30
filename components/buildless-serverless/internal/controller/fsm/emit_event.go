package fsm

import (
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	warningMessagePrefix = "Warning"
)

func emitEvent(m *StateMachine) {
	for _, condition := range m.State.Function.Status.Conditions {
		memorizedCondition := meta.FindStatusCondition(m.State.statusSnapshot.Conditions, condition.Type)

		if memorizedCondition != nil &&
			memorizedCondition.Status == condition.Status &&
			memorizedCondition.Reason == condition.Reason &&
			memorizedCondition.Message == condition.Message {
			continue
		}

		m.EventRecorder.Event(
			&m.State.Function,
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
