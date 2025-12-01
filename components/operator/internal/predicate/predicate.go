package predicate

import (
	"reflect"

	"sigs.k8s.io/controller-runtime/pkg/client"
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

	return isAnnotationUpdate(e) || !isStatusUpdate(e)
}

func isStatusUpdate(e event.UpdateEvent) bool {
	if e.ObjectOld.GetGeneration() == e.ObjectNew.GetGeneration() &&
		e.ObjectOld.GetResourceVersion() != e.ObjectNew.GetResourceVersion() {
		return true
	}

	return false
}

func isAnnotationUpdate(e event.UpdateEvent) bool {
	oldAnnotations := e.ObjectOld.GetAnnotations()
	newAnnotations := e.ObjectNew.GetAnnotations()
	return !reflect.DeepEqual(oldAnnotations, newAnnotations)
}

type ExactLabelPredicate struct {
	expectedKey   string
	expectedValue string
	predicate.Funcs
}

func NewExactLabelPredicate(key, value string) ExactLabelPredicate {
	return ExactLabelPredicate{
		expectedKey:   key,
		expectedValue: value,
	}
}

func (p ExactLabelPredicate) Create(e event.CreateEvent) bool {
	return isObjectExactLabel(e.Object, p.expectedKey, p.expectedValue)
}

func (p ExactLabelPredicate) Update(e event.UpdateEvent) bool {
	return isObjectExactLabel(e.ObjectNew, p.expectedKey, p.expectedValue)
}

func (p ExactLabelPredicate) Delete(e event.DeleteEvent) bool {
	return isObjectExactLabel(e.Object, p.expectedKey, p.expectedValue)
}

func (p ExactLabelPredicate) Generic(e event.GenericEvent) bool {
	return isObjectExactLabel(e.Object, p.expectedKey, p.expectedValue)
}

func isObjectExactLabel(obj client.Object, key, value string) bool {
	if obj == nil {
		return false
	}

	return obj.GetLabels() != nil && obj.GetLabels()[key] == value
}
