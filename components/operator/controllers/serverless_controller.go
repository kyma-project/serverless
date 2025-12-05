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

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/components/operator/internal/chart"
	"github.com/kyma-project/serverless/components/operator/internal/predicate"
	"github.com/kyma-project/serverless/components/operator/internal/state"
	"github.com/kyma-project/serverless/components/operator/internal/tracing"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

// serverlessReconciler reconciles a Serverless object
type serverlessReconciler struct {
	initStateMachine func(*zap.SugaredLogger) state.StateReconciler
	client           client.Client
	log              *zap.SugaredLogger
}

func NewServerlessReconciler(client client.Client, config *rest.Config, recorder record.EventRecorder, log *zap.SugaredLogger, chartPath string) *serverlessReconciler {
	cache := chart.NewSecretManifestCache(client)

	return &serverlessReconciler{
		initStateMachine: func(log *zap.SugaredLogger) state.StateReconciler {
			return state.NewMachine(client, config, recorder, log, cache, chartPath)
		},
		client: client,
		log:    log,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (sr *serverlessReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Serverless{}, builder.WithPredicates(predicate.NoStatusChangePredicate{})).
		Watches(&v1alpha1.Serverless{}, &handler.Funcs{
			// retrigger all Serverless CRs reconciliations when one is deleted
			// this should ensure at least one Serverless CR is served
			DeleteFunc: sr.retriggerAllServerlessCRsOnDelete,
		}).
		Watches(&corev1.Service{}, tracing.ServiceCollectorWatcher()).
		Watches(&appsv1.Deployment{}, &handler.Funcs{
			// retrigger all Serverless CRs reconciliations when a serverless-controller Deployment is updated or deleted
			UpdateFunc: sr.retriggerAllServerlessCRsOnUpdate,
			DeleteFunc: sr.retriggerAllServerlessCRsOnDelete,
		}, builder.WithPredicates(predicate.NewExactLabelPredicate("app.kubernetes.io/managed-by", "serverless-operator"))).
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

	//TODO: This is temporary solution, remove it after removing legacy serverless
	if err = sr.cleanupSvsCR(ctx, instance); err != nil {
		log.Warnf("failed to remove legacy serverless status fields: %s", err.Error())
	}

	r := sr.initStateMachine(log)
	return r.Reconcile(ctx, *instance)
}

// TODO: This is temporary solution, remove it after removing legacy serverless
func (sr *serverlessReconciler) cleanupSvsCR(ctx context.Context, instance *v1alpha1.Serverless) error {
	// remove deprecated status fields if present, unless buildless mode is disabled
	if mode, ok := instance.Annotations["serverless.kyma-project.io/buildless-mode"]; !ok || mode != "disabled" {
		return sr.removeLegacyStatusFields(ctx, types.NamespacedName{
			Namespace: instance.GetNamespace(),
			Name:      instance.GetName(),
		})
	}
	return nil
}

// TODO: This is temporary solution, remove it after removing legacy serverless
func (sr *serverlessReconciler) removeLegacyStatusFields(ctx context.Context, nn types.NamespacedName) error {
	// Build a JSON merge patch that nulls the deprecated fields.
	patchObj := map[string]any{
		"status": map[string]any{
			"defaultBuildJobPreset": nil,
			"dockerRegistry":        nil,
		},
	}
	patchBytes, err := json.Marshal(patchObj)
	if err != nil {
		return err
	}
	
	s := &v1alpha1.Serverless{}
	s.SetNamespace(nn.Namespace)
	s.SetName(nn.Name)
	return sr.client.Status().Patch(ctx, s, client.RawPatch(types.MergePatchType, patchBytes))
}

func (sr *serverlessReconciler) retriggerAllServerlessCRsOnUpdate(ctx context.Context, _ event.TypedUpdateEvent[client.Object], q workqueue.TypedRateLimitingInterface[ctrl.Request]) {
	sr.retriggerAllServerlessCRs(ctx, q)
}

func (sr *serverlessReconciler) retriggerAllServerlessCRsOnDelete(ctx context.Context, _ event.TypedDeleteEvent[client.Object], q workqueue.TypedRateLimitingInterface[ctrl.Request]) {
	sr.retriggerAllServerlessCRs(ctx, q)
}

func (sr *serverlessReconciler) retriggerAllServerlessCRs(ctx context.Context, q workqueue.TypedRateLimitingInterface[ctrl.Request]) {
	log := sr.log.With("deletion_watcher")

	list := &v1alpha1.ServerlessList{}
	err := sr.client.List(ctx, list, &client.ListOptions{})
	if err != nil {
		log.Errorf("error listing serverless objects: %s", err.Error())
		return
	}

	for _, s := range list.Items {
		log.Debugf("retriggering reconciliation for Serverless %s/%s", s.GetNamespace(), s.GetName())
		q.Add(ctrl.Request{NamespacedName: client.ObjectKey{
			Namespace: s.GetNamespace(),
			Name:      s.GetName(),
		}})
	}
}
