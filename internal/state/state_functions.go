package state

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kyma-project/module-manager/pkg/manifest"
	"github.com/kyma-project/module-manager/pkg/types"
	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/dependencies"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	requeueDuration = time.Second * 3
)

// choose right scenario to start (installation/deletion)
func sFnInitialize(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	// TODO: move defaulting to dedicated space
	s.instance.Spec.Default()

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
		// set addresses if they changed
		if s.instance.Status.PublisherProxyStatus != *s.instance.Spec.PublisherProxy.Value && s.instance.Status.TraceCollectorStatus != *s.instance.Spec.TraceCollector.Value {
			s.instance.Status.PublisherProxyStatus = *s.instance.Spec.PublisherProxy.Value
			s.instance.Status.TraceCollectorStatus = *s.instance.Spec.TraceCollector.Value
			sFnUpdateServerlessStatus(v1alpha1.StateProcessing)
		}

		// ping optional endpoints before installation

		//for _, URL := range []string{*s.instance.Spec.TraceCollector.Value, *s.instance.Spec.PublisherProxy.Value} {
		//	if err := dependencies.CheckOptionalDependencies(ctx, r.client, URL); err != nil {
		//		return sFnUpdateServerlessStatus(v1alpha1.StateError) //TODO : Moze inny state (Czy on musi blokowac?)
		//	}
		//}

		// run verification procedure if component is already installed
		if s.instance.Status.State == v1alpha1.StateReady {
			return sFnVerifyResources()
		}

		// when we know that cluster configuration met serverless requirements we can go to installation state
		return sFnApplyResources()
	}, nil, nil
}

// run serverless chart installation
func sFnApplyResources() (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		installationSpec, err := installationSpec(r, s)
		if err != nil {
			r.log.Errorf("error while preparing installation spec: %s", err.Error())
			return sFnUpdateServerlessStatus(v1alpha1.StateError)
		}

		installInfo, err := installInfo(ctx, r, s, installationSpec)
		if err != nil {
			r.log.Errorf("error while preparing installation info: %s", err.Error())
			return sFnUpdateServerlessStatus(v1alpha1.StateError)
		}

		ready, err := manifest.InstallChart(log.FromContext(ctx), installInfo, nil, r.cacheManager.GetRendererCache())
		if err != nil {
			r.log.Warnf("error while installing resource %s: %s",
				client.ObjectKeyFromObject(&s.instance), err.Error())
			return sFnUpdateServerlessStatus(v1alpha1.StateError)
		}
		if ready {
			return sFnUpdateServerlessStatus(v1alpha1.StateReady)
		}

		return requeueAfter(requeueDuration)
	}, nil, nil
}

// verify if installed component is up to date
func sFnVerifyResources() (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		installationSpec, err := installationSpec(r, s)
		if err != nil {
			r.log.Errorf("error while preparing installation spec: %s", err.Error())
			return sFnUpdateServerlessStatus(v1alpha1.StateError)
		}

		installInfo, err := installInfo(ctx, r, s, installationSpec)
		if err != nil {
			r.log.Errorf("error while preparing installation info: %s", err.Error())
			return sFnUpdateServerlessStatus(v1alpha1.StateError)
		}

		// verify installed resources
		ready, err := manifest.ConsistencyCheck(log.FromContext(ctx), installInfo, nil, r.cacheManager.GetRendererCache())

		// update only if resources not ready OR an error occurred during chart verification
		if err != nil {
			r.log.Errorf("error while checking resource %s: %s",
				client.ObjectKeyFromObject(&s.instance), err.Error())
			return sFnUpdateServerlessStatus(v1alpha1.StateError)
		} else if !ready {
			return sFnUpdateServerlessStatus(v1alpha1.StateProcessing)
		}

		return requeueAfter(requeueDuration)
	}, nil, nil
}

// delete installed resources
func sFnDeleteResources() (stateFn, *ctrl.Result, error) {
	return func(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
		// check if instance is in right state
		if s.instance.Status.State != v1alpha1.StateDeleting {
			return sFnUpdateServerlessStatus(v1alpha1.StateDeleting)
		}

		installationSpec, err := installationSpec(r, s)
		if err != nil {
			r.log.Errorf("can't prepare installation spec: %s", err.Error())
			return sFnUpdateServerlessStatus(v1alpha1.StateError)
		}

		// fallback logic for flags
		if installationSpec.SetFlags == nil {
			installationSpec.SetFlags = map[string]interface{}{}
		}
		if installationSpec.ConfigFlags == nil {
			installationSpec.ConfigFlags = map[string]interface{}{}
		}

		installInfo, err := installInfo(ctx, r, s, installationSpec)
		if err != nil {
			r.log.Errorf("can't prepare installation info: %s", err.Error())
			return sFnUpdateServerlessStatus(v1alpha1.StateError)
		}

		readyToBeDeleted, err := manifest.UninstallChart(log.FromContext(ctx), installInfo, nil, r.cacheManager.GetRendererCache())
		if err != nil {
			r.log.Warnf("error while deleting resource %s: %s",
				client.ObjectKeyFromObject(&s.instance), err.Error())
			return stopWithError(err)
		}

		// if resources are ready to be deleted, remove finalizer
		if readyToBeDeleted && controllerutil.RemoveFinalizer(&s.instance, r.finalizer) {
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

func installationSpec(r *reconciler, s *systemState) (*types.InstallationSpec, error) {
	flags, err := structToFlags(s.instance.Spec)
	if err != nil {
		return nil, fmt.Errorf("resolving manifest failed: %w", err)
	}

	// fetch install information
	installSpec := &types.InstallationSpec{
		ChartPath: r.chartPath,
		ChartFlags: types.ChartFlags{
			ConfigFlags: types.Flags{
				"Namespace":       r.chartNs,
				"CreateNamespace": r.createNamespace,
			},
			SetFlags: flags,
		},
	}
	if installSpec.ChartPath == "" {
		return nil, fmt.Errorf("no chart path available for processing")
	}

	return installSpec, nil
}

func installInfo(ctx context.Context, r *reconciler, s *systemState, installSpec *types.InstallationSpec) (types.InstallInfo, error) {
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&s.instance)
	if err != nil {
		return types.InstallInfo{}, err
	}

	return types.InstallInfo{
		Ctx: ctx,
		ChartInfo: &types.ChartInfo{
			ChartPath:   installSpec.ChartPath,
			ReleaseName: s.instance.GetName(),
			Flags:       installSpec.ChartFlags,
		},
		ClusterInfo: types.ClusterInfo{
			// destination cluster rest config
			Config: r.config,
			// destination cluster rest client
			Client: r.client,
		},
		ResourceInfo: types.ResourceInfo{
			// base operator resource to be passed for custom checks
			BaseResource: &unstructured.Unstructured{
				Object: unstructuredObj,
			},
		},
		CheckReadyStates: true,
	}, nil
}

func structToFlags(obj interface{}) (flags types.Flags, err error) {
	data, err := json.Marshal(obj)

	if err != nil {
		return
	}

	err = json.Unmarshal(data, &flags)
	return
}
