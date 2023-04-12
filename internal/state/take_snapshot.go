package state

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
)

var sFnTakePreInitSnapshot = buildSFnTakeSnapshot(sFnInitialize, nil, nil)

func buildSFnTakeSnapshot(next stateFn, result *ctrl.Result, err error) stateFn {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		s.saveServerlessStatus()
		return next, result, err
	}
}
