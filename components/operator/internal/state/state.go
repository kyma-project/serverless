package state

import (
	"time"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
)

var requeueResult = &ctrl.Result{
	Requeue: true,
}

func nextState(next stateFn) (stateFn, *ctrl.Result, error) {
	return next, nil, nil
}

func stopWithEventualError(err error) (stateFn, *ctrl.Result, error) {
	return nil, nil, err
}

func stop() (stateFn, *ctrl.Result, error) {
	return nil, nil, nil
}

func requeue() (stateFn, *ctrl.Result, error) {
	return nil, requeueResult, nil
}

func requeueAfter(duration time.Duration) (stateFn, *ctrl.Result, error) {
	return nil, &ctrl.Result{
		RequeueAfter: duration,
	}, nil
}

type fieldsToUpdate []struct {
	specField    string
	statusField  *string
	fieldName    string
	defaultValue string
}

func updateStatusFields(eventRecorder record.EventRecorder, instance *v1alpha1.Serverless, fields fieldsToUpdate) {
	for _, field := range fields {
		// set default value if spec field is empty
		if field.specField == "" {
			field.specField = field.defaultValue
		}

		if field.specField != *field.statusField {
			oldStatusValue := *field.statusField
			*field.statusField = field.specField
			eventRecorder.Eventf(
				instance,
				"Normal",
				string(v1alpha1.ConditionReasonConfiguration),
				"%s set from '%s' to '%s'",
				field.fieldName,
				oldStatusValue,
				field.specField,
			)
		}
	}
}
