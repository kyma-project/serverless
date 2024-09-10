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
	client client.Client
	logger *logrus.Logger
}

func NewFunctionReconciler(client client.Client, logger *logrus.Logger) *FunctionReconciler {
	return &FunctionReconciler{
		client: client,
		logger: logger,
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



	// TODO: update



	// TODO: delete

	

	return reconcile.Result{}, nil
}
