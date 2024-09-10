package controllers

import (
	"context"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/pkg/apis/serverless/v1alpha2"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type FunctionReconciler struct {
	client              client.Client
	logger              *logrus.Logger
	functionSourcesPath string
}

func NewFunctionReconciler(client client.Client, logger *logrus.Logger, functionSourcesPath string) *FunctionReconciler {
	return &FunctionReconciler{
		client:              client,
		logger:              logger,
		functionSourcesPath: functionSourcesPath,
	}
}

func (r *FunctionReconciler) SetupWithManager(mgr manager.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("function-controller").
		For(&serverlessv1alpha2.Function{}, builder.WithPredicates(predicate.Funcs{UpdateFunc: IsNotFunctionStatusUpdate(r.logger)})).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

func (r *FunctionReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	r.logger.Info("starting reconciliation for", request.NamespacedName)

	// TODO: create

	// 1. get function from cluster
	// 2. set finalizer
	// 3.1. save sources to pvc
	// 3.2. ??? run npm install
	// 4. create deploy & service

	// TODO: update

	// 1. get function from cluster
	// 2.1. update sources
	// 2.2. ??? rerun npm install
	// 3. ??? update deploy ( we should have hot-reload enabled on every runtime )

	// TODO: delete

	// 1. remove deploy & service
	// 2. cleanup pvc
	// 3. remove finalizer

	return reconcile.Result{}, nil
}
