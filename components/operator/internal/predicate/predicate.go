package predicate

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// this predicate allows not reacting on status changes
type NoStatusChangePredicate struct {
	predicate.Funcs
}

func (p NoStatusChangePredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectNew == nil || e.ObjectOld == nil {
		return false
	}

	// first resource version (after apply)
	if e.ObjectOld.GetResourceVersion() == e.ObjectNew.GetResourceVersion() {
		return true
	}

	return !isStatusUpdate(e)
}

func isStatusUpdate(e event.UpdateEvent) bool {
	if e.ObjectOld.GetGeneration() == e.ObjectNew.GetGeneration() &&
		e.ObjectOld.GetResourceVersion() != e.ObjectNew.GetResourceVersion() {
		return true
	}

	return false
}
