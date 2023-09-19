package state

import (
	"context"
	"reflect"
	"time"
)

var (
	requeueDuration = time.Second * 3
)

func updateServerlessWithoutStatus(ctx context.Context, r *reconciler, s *systemState) error {
	return r.client.Update(ctx, &s.instance)
}

func updateServerlessStatus(ctx context.Context, r *reconciler, s *systemState) error {
	if !reflect.DeepEqual(s.instance.Status, s.statusSnapshot) {
		err := r.client.Status().Update(ctx, &s.instance)
		emitEvent(r, s)
		s.saveStatusSnapshot()
		s.instance.Spec.Default()
		return err
	}
	return nil
}
