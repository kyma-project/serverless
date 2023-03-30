package state

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	requeueDuration = time.Second * 3
)

func sFnUpdateServerlessStatus(state v1alpha1.State) (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		s.setState(state)
		err := r.client.Status().Update(ctx, &s.instance)
		if err != nil {
			stopWithError(err)
		}
		return requeue()
	}, nil, nil
}

func sFnUpdateServerless() (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		return nil, nil, r.client.Update(ctx, &s.instance)
	}, nil, nil
}

func stopWithError(err error) (stateFn, *ctrl.Result, error) {
	return nil, nil, err
}

func stop() (stateFn, *ctrl.Result, error) {
	return nil, nil, nil
}

func requeue() (stateFn, *ctrl.Result, error) {
	return nil, &ctrl.Result{
		Requeue: true,
	}, nil
}

func requeueAfter(duration time.Duration) (stateFn, *ctrl.Result, error) {
	return nil, &ctrl.Result{
		RequeueAfter: duration,
	}, nil
}

func chartConfig(ctx context.Context, r *reconciler, s *systemState) (*chart.Config, error) {
	flags, err := structToFlags(s)
	if err != nil {
		return nil, fmt.Errorf("resolving manifest failed: %w", err)
	}

	return &chart.Config{
		Ctx:   ctx,
		Log:   r.log,
		Cache: r.cache,
		Cluster: chart.Cluster{
			Client: r.client,
			Config: r.config,
		},
		Release: chart.Release{
			Flags:     flags,
			ChartPath: r.chartPath,
			Namespace: s.instance.GetNamespace(),
			Name:      s.instance.GetName(),
		},
	}, nil
}

func structToFlags(s *systemState) (flags map[string]interface{}, err error) {
	data, err := json.Marshal(s.instance.Spec)

	if err != nil {
		return
	}
	err = json.Unmarshal(data, &flags)

	enrichFlagsWithDockerRegistry(s.dockerRegistry, &flags)
	return
}

func enrichFlagsWithDockerRegistry(d map[string]interface{}, flags *map[string]interface{}) {
	newDockerRegistry := (*flags)["dockerRegistry"].(map[string]interface{})
	for _, k := range []string{"username", "password", "registryAddress", "serverAddress"} {
		if v, ok := d[k]; ok {
			newDockerRegistry[k] = v
		}
	}
	(*flags)["dockerRegistry"] = newDockerRegistry
}
