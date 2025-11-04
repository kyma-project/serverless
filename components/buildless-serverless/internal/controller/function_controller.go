/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"errors"
	"time"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/config"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/cache"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/fsm"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/state"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const healthCheckTimeout = time.Second

// FunctionReconciler reconciles a Function object
type FunctionReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	Log             *zap.SugaredLogger
	Config          config.FunctionConfig
	LastCommitCache cache.Cache
	EventRecorder   record.EventRecorder
	HealthCh        chan bool
}

// +kubebuilder:rbac:groups=serverless.kyma-project.io,resources=functions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=serverless.kyma-project.io,resources=functions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;update;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// TODO: This is temporary, it is necessary to delete orphaned resources
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=list;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=list;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=list;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (fr *FunctionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if IsHealthCheckRequest(req) {
		fr.sendHealthCheck()
		return ctrl.Result{}, nil
	}

	log := fr.Log.With("request", req)
	log.Info("reconciliation started")

	var instance serverlessv1alpha2.Function
	if err := fr.Get(ctx, req.NamespacedName, &instance); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if !instance.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	sm := fsm.New(fr.Client, fr.Config, &instance, state.StartState(), fr.EventRecorder, fr.Scheme, fr.LastCommitCache, log)
	return sm.Reconcile(ctx)
}

// SetupWithManager sets up the controller with the Manager.
func (fr *FunctionReconciler) SetupWithManager(mgr ctrl.Manager) (controller.Controller, error) {
	return ctrl.NewControllerManagedBy(mgr).
		Named("function-controller").
		For(&serverlessv1alpha2.Function{}).
		WithEventFilter(buildPredicates()).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		WithOptions(controller.Options{
			RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[reconcile.Request](
				5*time.Millisecond,
				1000*time.Second,
			),
		}).
		Build(fr)
}

func (fr *FunctionReconciler) sendHealthCheck() {
	fr.Log.Debug("health check request received")

	select {
	case fr.HealthCh <- true:
		fr.Log.Debug("health check request responded")
	case <-time.After(healthCheckTimeout):
		fr.Log.Warn(errors.New("timeout when responding to health check"))
	}
}

func buildPredicates() predicate.Funcs {
	// Predicate to skip reconciliation when the object is being deleted
	return predicate.Funcs{
		// Allow update events
		UpdateFunc: func(e event.UpdateEvent) bool {
			return true
		},
		// Allow create events
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		// Don't allow delete events
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		// Allow generic events (e.g., external triggers)
		GenericFunc: func(e event.GenericEvent) bool {
			return true
		},
	}
}
