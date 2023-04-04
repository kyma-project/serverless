package state

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnTakePreInitSnapshot(_ context.Context, _ *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	next, result, err := sFnTakeSnapsho(sFnInitialize, nil, nil)
	return next, result, err
}

func sFnTakeSnapsho(next stateFn, result *ctrl.Result, err error) (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		s.saveServerlessStatus()
		return next, result, err
	}, nil, nil

}
