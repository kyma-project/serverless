package state

import (
	"context"
	"reflect"
	"time"
)

var (
	requeueDuration = time.Second * 3
)

func updateDockerRegistryWithoutStatus(ctx context.Context, r *reconciler, s *systemState) error {
	return r.client.Update(ctx, &s.instance)
}

func updateDockerRegistryStatus(ctx context.Context, r *reconciler, s *systemState) error {
	if !reflect.DeepEqual(s.instance.Status, s.statusSnapshot) {
		err := r.client.Status().Update(ctx, &s.instance)
		emitEvent(r, s)
		s.saveStatusSnapshot()
		return err
	}
	return nil
}
