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
	"k8s.io/apimachinery/pkg/runtime/schema"
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
func (r *serverlessReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Serverless{}, builder.WithPredicates(predicate.NoStatusChangePredicate{})).
		Watches(&source.Kind{Type: &telemetryv1alpha1.TracePipeline{}}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}

func (sr *serverlessReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var instance v1alpha1.Serverless

	log := sr.log.With("request", req)
	log.Info("reconciliation started")

	if err := sr.client.Get(ctx, req.NamespacedName, &instance); err != nil {
		if k8serrors.IsNotFound(err) {
			//We are called because Watch was fired
			//TODO: refactor this
			var deeperErr error
			instance, deeperErr = sr.getServerless(ctx, log)
			if deeperErr != nil {
				log.Warnf("while getting serverless list, got error: %s", deeperErr.Error())
				return ctrl.Result{}, client.IgnoreNotFound(deeperErr)
			}
		} else {
			log.Warnf("while getting serverless, got error: %s", err.Error())
			return ctrl.Result{}, err
		}
	}

	r := sr.initStateMachine(log)
	return r.Reconcile(ctx, instance)
}

func (sr *serverlessReconciler) getServerless(ctx context.Context, log *zap.SugaredLogger) (v1alpha1.Serverless, error) {
	serverlesses := v1alpha1.ServerlessList{}
	err := sr.client.List(ctx, &serverlesses, &client.ListOptions{})
	if err != nil {
		return v1alpha1.Serverless{}, errors.Wrap(err, "while getting list of serverlesses")
	}
	serverlessAmount := len(serverlesses.Items)
	if serverlessAmount > 1 {
		log.Warn("Cluster has more than one serverless")
		return v1alpha1.Serverless{}, nil //TODO: think what to do in such scenario
	} else if serverlessAmount == 0 {
		//Is it possible? yes
		return v1alpha1.Serverless{}, k8serrors.NewNotFound(schema.GroupResource{
			Group:    v1alpha1.ServerlessGroup,
			Resource: v1alpha1.ServerlessKind,
		}, "")
	}
	return serverlesses.Items[0], nil
}
