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
	"time"

	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	"github.com/kyma-project/serverless-manager/internal/state"
)

// serverlessReconciler reconciles a Serverless object
type serverlessReconciler struct {
	initStateMachine func(*zap.SugaredLogger) state.StateReconciler
	client           client.Client
	log              *zap.SugaredLogger
}

func NewServerlessReconciler(client client.Client, config *rest.Config, log *zap.SugaredLogger, chartPath string) *serverlessReconciler {
	cache := chart.NewManifestCache()

	return &serverlessReconciler{
		initStateMachine: func(log *zap.SugaredLogger) state.StateReconciler {
			return state.NewMachine(client, config, log, cache, chartPath)
		},
		client: client,
		log:    log,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *serverlessReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Serverless{}).
		Complete(r)
}

func (sr *serverlessReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var instance v1alpha1.Serverless

	log := sr.log.With("request", req)
	log.Info("reconciliation started")

	if err := sr.client.Get(ctx, req.NamespacedName, &instance); err != nil {
		return ctrl.Result{
			RequeueAfter: time.Second * 30,
		}, client.IgnoreNotFound(err)
	}

	r := sr.initStateMachine(log)
	return r.Reconcile(ctx, instance)
}
