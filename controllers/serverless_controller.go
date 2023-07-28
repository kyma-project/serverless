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
	"github.com/kyma-project/serverless-manager/internal/tools"
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
func (sr *serverlessReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Serverless{}, builder.WithPredicates(predicate.NoStatusChangePredicate{})).
		Watches(&source.Kind{Type: &telemetryv1alpha1.TracePipeline{}}, &handler.EnqueueRequestForObject{}).
		Complete(sr)
}

func (sr *serverlessReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var instance v1alpha1.Serverless

	log := sr.log.With("request", req)
	log.Info("reconciliation started")

	var err error
	instance, err = sr.getServerless(ctx, log)
	if err != nil {
		log.Warnf("while getting serverless, got error: %s", err.Error())
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	r := sr.initStateMachine(log)
	return r.Reconcile(ctx, instance)
}

/*
This method list all serverless in cluster, because reconcile loop can be launched with request created by watch
*/
func (sr *serverlessReconciler) getServerless(ctx context.Context, log *zap.SugaredLogger) (v1alpha1.Serverless, error) {
	serverlessList := v1alpha1.ServerlessList{}
	err := sr.client.List(ctx, &serverlessList, &client.ListOptions{})
	if err != nil {
		return v1alpha1.Serverless{}, errors.Wrap(err, "while getting list of serverlessList")
	}
	serverlessAmount := len(serverlessList.Items)
	if serverlessAmount == 0 {
		return v1alpha1.Serverless{}, k8serrors.NewNotFound(schema.GroupResource{
			Group:    v1alpha1.ServerlessGroup,
			Resource: v1alpha1.ServerlessKind,
		}, "")
	} else if serverlessAmount > 1 {
		log.Warn("Cluster has more than one serverless")
		err := updateAllServerlessInstancesWithError(ctx, sr.client, serverlessList)
		return v1alpha1.Serverless{}, errors.Wrap(err, "while setting errors on all serverless instances.")
	}

	return serverlessList.Items[0], nil
}

func updateAllServerlessInstancesWithError(ctx context.Context, c client.Client, list v1alpha1.ServerlessList) error {
	errList := tools.NewErrorList()
	for _, item := range list.Items {
		item.Status.State = v1alpha1.StateError
		item.UpdateConditionFalse(
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonServerlessDuplicated,
			errors.New("Found more than one Serverless CR. To fix please remove redundant serverless CRs"),
		)
		err := c.Update(ctx, &item, &client.UpdateOptions{})
		errList.Append(err)
	}
	return errList.ToError()
}
