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
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	"github.com/kyma-project/serverless-manager/internal/predicate"
	"github.com/kyma-project/serverless-manager/internal/state"
	telemetryv1alpha1 "github.com/kyma-project/telemetry-manager/apis/telemetry/v1alpha1"
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
		Watches(&source.Kind{Type: &telemetryv1alpha1.TracePipeline{}}, &handler.EnqueueRequestForObject{}).
		Complete(sr)
}

func (sr *serverlessReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := sr.log.With("request", req)
	log.Info("reconciliation started")

	instance, err := sr.getServerless(ctx, req)
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

func (sr *serverlessReconciler) getServerless(ctx context.Context, req ctrl.Request) (*v1alpha1.Serverless, error) {
	instance := &v1alpha1.Serverless{}
	err := sr.client.Get(ctx, req.NamespacedName, instance)
	if err == nil {
		return instance, nil
	}
	if !k8serrors.IsNotFound(err) {
		return nil, errors.Wrap(err, "while fetching serverless instance")
	}

	instance, err = findServedServerless(ctx, sr.client)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching served serverless instance")
	}
	return instance, nil
}

func findServedServerless(ctx context.Context, c client.Client) (*v1alpha1.Serverless, error) {
	var serverlessList v1alpha1.ServerlessList

	err := c.List(ctx, &serverlessList)

	if err != nil {
		return nil, err
	}

	for _, item := range serverlessList.Items {
		if !item.IsServedEmpty() && item.Status.Served == v1alpha1.ServedTrue {
			return &item, nil
		}
	}

	return nil, nil
}
