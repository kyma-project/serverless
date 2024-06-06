package serverless

import (
	ctrl "sigs.k8s.io/controller-runtime"
)

func NextRequeue(err error) (res ctrl.Result, errMsg string) {
	// use exponential delay
	return ctrl.Result{Requeue: true}, errMsg
}
