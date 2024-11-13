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
	"github.com/go-logr/logr"
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	pkglog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"
)

// FunctionReconciler reconciles a Function object
type FunctionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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
	log := pkglog.FromContext(ctx).WithName(req.NamespacedName.Name)

	log.Info("reconciliation started")

	var function serverlessv1alpha2.Function
	if err := r.Get(ctx, req.NamespacedName, &function); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("function spec", "foo", function.Spec.Foo)

	return r.handleDeployment(ctx, function, log)
}

func (r *FunctionReconciler) handleDeployment(ctx context.Context, function serverlessv1alpha2.Function, log logr.Logger) (ctrl.Result, error) {
	newDeployment := r.constructDeploymentForFunction(&function)

	currentDeployment := &appsv1.Deployment{}
	deploymentErr := r.Get(ctx, client.ObjectKey{
		Namespace: function.Namespace,
		Name:      fmt.Sprintf("%s-function-deployment", function.Name),
	}, currentDeployment)
	if deploymentErr != nil && apierrors.IsNotFound(deploymentErr) {
		log.Info("creating a new Deployment", "Deployment.Namespace", newDeployment.Namespace, "Deployment.Name", newDeployment.Name)
		if err := r.Create(ctx, newDeployment); err != nil {
			log.Error(err, "failed to create new Deployment", "Deployment.Namespace", newDeployment.Namespace, "Deployment.Name", newDeployment.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	} else if deploymentErr != nil {
		log.Error(deploymentErr, "unable to fetch Deployment for Function")
		return ctrl.Result{}, deploymentErr
	}

	result, err := r.updateDeploymentIfNeeded(ctx, currentDeployment, newDeployment, log)
	if err != nil {
		return result, err
	}
	return ctrl.Result{}, nil
}

func (r *FunctionReconciler) updateDeploymentIfNeeded(ctx context.Context, currentDeployment *appsv1.Deployment, newDeployment *appsv1.Deployment, log logr.Logger) (ctrl.Result, error) {
	// Ensure the Deployment data matches the desired state
	if currentDeployment.Spec.Template.Annotations["foo"] != newDeployment.Spec.Template.Annotations["foo"] {
		currentDeployment.Spec.Template.Annotations["foo"] = newDeployment.Spec.Template.Annotations["foo"]
		if err := r.Update(ctx, currentDeployment); err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", currentDeployment.Namespace, "Deployment.Name", currentDeployment.Name)
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

func (r *FunctionReconciler) constructDeploymentForFunction(function *serverlessv1alpha2.Function) *appsv1.Deployment {
	labels := map[string]string{
		"app": function.Name,
		//"foo": function.Spec.Foo,
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-function-deployment", function.Name),
			Namespace: function.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
					Annotations: map[string]string{
						"foo": function.Spec.Foo,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  fmt.Sprintf("%s-function-pod", function.Name),
							Image: "nginx:latest",
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	// Set the ownerRef for the Deployment, ensuring that the Deployment
	// will be deleted when the Busybox CR is deleted.
	controllerutil.SetControllerReference(function, deployment, r.Scheme)

	return deployment
}
