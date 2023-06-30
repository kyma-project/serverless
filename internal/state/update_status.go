package state

import (
	"context"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	requeueDuration = time.Second * 3
)

func sFnUpdateStatusAndRequeue(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	err := updateServerlessStatus(ctx, r, s)
	return stopWithErrorOrRequeue(err)
}

func sFnUpdateStatusAndStop(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	err := updateServerlessStatus(ctx, r, s)
	return stopWithError(err)
}

// TODO: can we refactor this?
func sFnUpdateServerless(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	return stopWithError(r.client.Update(ctx, &s.instance))
}

func updateServerlessStatus(ctx context.Context, r *reconciler, s *systemState) error {
	instance := s.instance.DeepCopy()
	err := r.client.Status().Update(ctx, instance)
	emitEvent(r, s)
	return err
}
