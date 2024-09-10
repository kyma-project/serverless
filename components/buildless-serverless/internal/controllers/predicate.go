package controllers

import (
	"reflect"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/pkg/apis/serverless/v1alpha2"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func IsNotFunctionStatusUpdate(log *logrus.Logger) func(event.UpdateEvent) bool {
	return func(event event.UpdateEvent) bool {
		if event.ObjectOld == nil || event.ObjectNew == nil {
			return true
		}

		log.Debug("old: ", event.ObjectOld.GetName())
		log.Debug("new: ", event.ObjectNew.GetName())

		oldFn, ok := event.ObjectOld.(*serverlessv1alpha2.Function)
		if !ok {
			v := reflect.ValueOf(event.ObjectOld)
			log.Debug("Can't cast to function from type: ", v.Type())
			return true
		}

		newFn, ok := event.ObjectNew.(*serverlessv1alpha2.Function)
		if !ok {
			v := reflect.ValueOf(event.ObjectNew)
			log.Debug("Can't cast to function from type: ", v.Type())
			return true
		}

		equalStatus := equalFunctionStatus(oldFn.Status, newFn.Status)
		log.Debug("Statuses are equal: ", equalStatus)

		return equalStatus
	}
}

func equalFunctionStatus(left, right serverlessv1alpha2.FunctionStatus) bool {
	if !equalConditions(left.Conditions, right.Conditions) {
		return false
	}

	if left.Repository != right.Repository ||
		left.Commit != right.Commit ||
		left.Runtime != right.Runtime {
		return false
	}
	return true
}

func equalConditions(existing, expected []serverlessv1alpha2.Condition) bool {
	if len(existing) != len(expected) {
		return false
	}

	existingMap := make(map[serverlessv1alpha2.ConditionType]serverlessv1alpha2.Condition, len(existing))
	for _, value := range existing {
		existingMap[value.Type] = value
	}

	for _, expectedCondition := range expected {
		existingCondition := existingMap[expectedCondition.Type]
		if !existingCondition.Equal(&expectedCondition) {
			return false
		}
	}
	return true
}
