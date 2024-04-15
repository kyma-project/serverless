package state

import (
	"context"
	"fmt"
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
		r.log.Debug(fmt.Sprintf("updating serverless status to '%+v'", s.instance.Status))
		err := r.client.Status().Update(ctx, &s.instance)
		emitEvent(r, s)
		s.saveStatusSnapshot()
		return err
	}
	return nil
}
