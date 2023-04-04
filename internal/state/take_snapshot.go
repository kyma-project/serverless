package state

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnTakeSnapshot(_ context.Context, _ *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	s.saveServerlessStatus()
	return sFnInitialize, nil, nil
}
