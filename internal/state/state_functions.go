package state

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	"github.com/kyma-project/serverless-manager/internal/dependencies"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var (
	requeueDuration = time.Second * 3
)

// choose right scenario to start (installation/deletion)
func sFnInitialize(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	instanceIsBeingDeleted := !s.instance.GetDeletionTimestamp().IsZero()
	instanceHasFinalizer := controllerutil.ContainsFinalizer(&s.instance, r.finalizer)

	// in case instance does not have finalizer - add it and update instance
	if !instanceIsBeingDeleted && !instanceHasFinalizer {
		controllerutil.AddFinalizer(&s.instance, r.finalizer)
		err := r.client.Update(ctx, &s.instance)
		// stop state machine with potential error
		return stopWithError(err)
	}

	// in case instance has no finalizer and instance is being deleted - end reconciliation
	if instanceIsBeingDeleted && !instanceHasFinalizer {
		// stop state machine
		return stop()
	}

	if s.instance.Status.State.IsEmpty() {
		return sFnUpdateServerlessStatus(v1alpha1.StateProcessing)
	}

	err := s.Setup(ctx, r.client)
	if err != nil {
		return sFnUpdateServerlessStatus(v1alpha1.StateError)
	}

	// in case instance is being deleted and has finalizer - delete all resources
	if instanceIsBeingDeleted {
		return sFnDeleteResources()
	}

	return sFnPrerequisites()
}

// check necessery dependencies before installation
func sFnPrerequisites() (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		// check hard serverless dependencies before installation
		withIstio := s.instance.Spec.DockerRegistry.IsInternalEnabled()
		err := dependencies.CheckPrerequisites(ctx, r.client, withIstio)
		if err != nil {
			return sFnUpdateServerlessStatus(v1alpha1.StateError)
		}

		// when we know that cluster configuration met serverless requirements we can go to installation state
		return sFnApplyResources()
	}, nil, nil
}

// run serverless chart installation
func sFnApplyResources() (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		chartConfig, err := chartConfig(ctx, r, s)
		if err != nil {
			r.log.Errorf("error while preparing chart config: %s", err.Error())
			return sFnUpdateServerlessStatus(v1alpha1.StateError)
		}

		err = chart.Install(chartConfig)
		if err != nil {
			r.log.Warnf("error while installing resource %s: %s",
				client.ObjectKeyFromObject(&s.instance), err.Error())
			return sFnUpdateServerlessStatus(v1alpha1.StateError)
		}

		return sFnVerifyResources(chartConfig)
	}, nil, nil
}

func sFnVerifyResources(chartConfig *chart.Config) (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		ready, err := chart.Verify(chartConfig)
		if err != nil {
			r.log.Warnf("error while verifying resource %s: %s",
				client.ObjectKeyFromObject(&s.instance), err.Error())
			return sFnUpdateServerlessStatus(v1alpha1.StateError)
		}

		if ready {
			return sFnUpdateServerlessStatus(v1alpha1.StateReady)
		}

		return requeueAfter(requeueDuration)
	}, nil, nil
}

func sFnDeleteResources() (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		// check if instance is in right state
		if s.instance.Status.State != v1alpha1.StateDeleting {
			return sFnUpdateServerlessStatus(v1alpha1.StateDeleting)
		}

		chartConfig, err := chartConfig(ctx, r, s)
		if err != nil {
			r.log.Errorf("error while preparing chart config: %s", err.Error())
			return sFnUpdateServerlessStatus(v1alpha1.StateError)
		}

		err = chart.Uninstall(chartConfig)
		if err != nil {
			r.log.Warnf("error while uninstalling resource %s: %s",
				client.ObjectKeyFromObject(&s.instance), err.Error())
			return sFnUpdateServerlessStatus(v1alpha1.StateError)
		}

		// if resources are ready to be deleted, remove finalizer
		if controllerutil.RemoveFinalizer(&s.instance, r.finalizer) {
			return sFnUpdateServerless()
		}

		return requeue()
	}, nil, nil
}

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
		Ctx:    ctx,
		Log:    r.log,
		Client: r.client,
		Cache:  r.cache,
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
