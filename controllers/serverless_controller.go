/*
Copyright 2022.

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

package controllers

import (
	"context"
	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	"github.com/kyma-project/serverless-manager/internal/predicate"
	"github.com/kyma-project/serverless-manager/internal/state"
	"github.com/kyma-project/serverless-manager/internal/tracing"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

// serverlessReconciler reconciles a Serverless object
type serverlessReconciler struct {
	initStateMachine func(*zap.SugaredLogger) state.StateReconciler
	client           client.Client
	log              *zap.SugaredLogger
}

func NewServerlessReconciler(client client.Client, config *rest.Config, recorder record.EventRecorder, log *zap.SugaredLogger, chartPath string, ns string) *serverlessReconciler {
	cache := chart.NewSecretManifestCache(client)

	return &serverlessReconciler{
		initStateMachine: func(log *zap.SugaredLogger) state.StateReconciler {
			return state.NewMachine(client, config, recorder, log, cache, chartPath, ns)
		},
		client: client,
		log:    log,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (sr *serverlessReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Serverless{}, builder.WithPredicates(predicate.NoStatusChangePredicate{})).
		Watches(&source.Kind{Type: &corev1.Service{}}, tracing.ServiceCollectorWatcher()).
		WithOptions(controller.Options{
			RateLimiter: workqueue.NewMaxOfRateLimiter(
				workqueue.NewItemExponentialFailureRateLimiter(3*time.Second, 300*time.Second),
				// 10 qps, 100 bucket size.  This is only for retry speed and its only the overall factor (not per item)
				&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
			),
			MaxConcurrentReconciles: 1, // Build job scheduling mechanism requires this parameter to be set to 1. The mechanism is based on getting active and stateless jobs, concurrent reconciles makes it non deterministic . Value 1 removes data races while fetching list of jobs. https://github.com/kyma-project/kyma/issues/10037
		}).
		Complete(sr)
}

func (sr *serverlessReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := sr.log.With("request", req)
	log.Info("reconciliation started")

	instance, err := state.GetServerlessOrServed(ctx, req, sr.client)
	if err != nil {
		log.Warnf("while getting serverless, got error: %s", err.Error())
		return ctrl.Result{}, errors.Wrap(err, "while fetching serverless instance")
	}
	if instance == nil {
		log.Info("Couldn't find proper instance of serverless")
		return ctrl.Result{}, nil
	}

	r := sr.initStateMachine(log)
	return r.Reconcile(ctx, *instance)
}
