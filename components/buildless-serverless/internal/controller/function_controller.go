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
	"fmt"
	"github.com/google/go-cmp/cmp"
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"
)

// FunctionReconciler reconciles a Function object
type FunctionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	//TODO: Add to logger request data somehow
	Log *zap.SugaredLogger
}

// +kubebuilder:rbac:groups=serverless.kyma-project.io,resources=functions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=serverless.kyma-project.io,resources=functions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=serverless.kyma-project.io,resources=functions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Function object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *FunctionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("reconciliation started")

	var function serverlessv1alpha2.Function
	if err := r.Get(ctx, req.NamespacedName, &function); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.handleDeployment(ctx, function)
}

func (r *FunctionReconciler) handleDeployment(ctx context.Context, function serverlessv1alpha2.Function) (ctrl.Result, error) {
	builtDeployment := buildDeployment(&function)

	clusterDeployment, resultGet, errGet := r.getOrCreateDeployment(ctx, function, builtDeployment)
	if clusterDeployment == nil {
		return resultGet, errGet
	}

	resultUpdate, errUpdate := r.updateDeploymentIfNeeded(ctx, clusterDeployment, builtDeployment)
	if errUpdate != nil {
		return resultUpdate, errUpdate
	}
	return ctrl.Result{}, nil
}

func (r *FunctionReconciler) getOrCreateDeployment(ctx context.Context, function serverlessv1alpha2.Function, builtDeployment *appsv1.Deployment) (*appsv1.Deployment, ctrl.Result, error) {
	currentDeployment := &appsv1.Deployment{}
	deploymentErr := r.Get(ctx, client.ObjectKey{
		Namespace: function.Namespace,
		Name:      fmt.Sprintf("%s-function-deployment", function.Name),
	}, currentDeployment)

	if deploymentErr == nil {
		return currentDeployment, ctrl.Result{}, nil
	}
	if !apierrors.IsNotFound(deploymentErr) {
		r.Log.Error(deploymentErr, "unable to fetch Deployment for Function")
		return nil, ctrl.Result{}, deploymentErr
	}

	createResult, createErr := r.createDeployment(ctx, function, builtDeployment)
	return nil, createResult, createErr
}

func (r *FunctionReconciler) createDeployment(ctx context.Context, function serverlessv1alpha2.Function, deployment *appsv1.Deployment) (ctrl.Result, error) {
	r.Log.Info("creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)

	// Set the ownerRef for the Deployment, ensuring that the Deployment
	// will be deleted when the Function CR is deleted.
	controllerutil.SetControllerReference(&function, deployment, r.Scheme)

	if err := r.Create(ctx, deployment); err != nil {
		r.Log.Error(err, "failed to create new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

func (r *FunctionReconciler) updateDeploymentIfNeeded(ctx context.Context, clusterDeployment *appsv1.Deployment, builtDeployment *appsv1.Deployment) (ctrl.Result, error) {
	// Ensure the Deployment data matches the desired state
	if !cmp.Equal(clusterDeployment.Spec.Template, builtDeployment.Spec.Template) {
		clusterDeployment.Spec.Template = builtDeployment.Spec.Template
		if err := r.Update(ctx, clusterDeployment); err != nil {
			r.Log.Error(err, "Failed to update Deployment", "Deployment.Namespace", clusterDeployment.Namespace, "Deployment.Name", clusterDeployment.Name)
			return ctrl.Result{}, err
		}
		// Requeue the request to ensure the Deployment is updated
		return ctrl.Result{Requeue: true}, nil
	}
	return ctrl.Result{}, nil
}

func buildPredicates() predicate.Funcs {
	// Predicate to skip reconciliation when the object is being deleted
	return predicate.Funcs{
		// Allow create events
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

// SetupWithManager sets up the controller with the Manager.
func (r *FunctionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&serverlessv1alpha2.Function{}).
		WithEventFilter(buildPredicates()).
		Owns(&appsv1.Deployment{}).
		Named("function").
		Complete(r)
}
