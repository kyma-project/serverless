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
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	r.Log.Info("function spec", "foo", function.Spec.Foo)

	return r.handleDeployment(ctx, function)
}

func (r *FunctionReconciler) handleDeployment(ctx context.Context, function serverlessv1alpha2.Function) (ctrl.Result, error) {
	newDeployment := r.constructDeploymentForFunction(&function)

	currentDeployment, resultGet, errGet := r.getDeployment(ctx, function, newDeployment)
	if currentDeployment == nil {
		return resultGet, errGet
	}

	resultUpdate, errUpdate := r.updateDeploymentIfNeeded(ctx, currentDeployment, newDeployment)
	if errUpdate != nil {
		return resultUpdate, errUpdate
	}
	return ctrl.Result{}, nil
}

func (r *FunctionReconciler) getDeployment(ctx context.Context, function serverlessv1alpha2.Function, newDeployment *appsv1.Deployment) (*appsv1.Deployment, ctrl.Result, error) {
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

	r.Log.Info("creating a new Deployment", "Deployment.Namespace", newDeployment.Namespace, "Deployment.Name", newDeployment.Name)
	if err := r.Create(ctx, newDeployment); err != nil {
		r.Log.Error(err, "failed to create new Deployment", "Deployment.Namespace", newDeployment.Namespace, "Deployment.Name", newDeployment.Name)
		return nil, ctrl.Result{}, err
	}
	return nil, ctrl.Result{RequeueAfter: time.Minute}, nil
}

func (r *FunctionReconciler) updateDeploymentIfNeeded(ctx context.Context, currentDeployment *appsv1.Deployment, newDeployment *appsv1.Deployment) (ctrl.Result, error) {
	// Ensure the Deployment data matches the desired state
	//TODO: write better if to react to changes (not to only Annotation "Foo")
	if currentDeployment.Spec.Template.Annotations["foo"] != newDeployment.Spec.Template.Annotations["foo"] {
		currentDeployment.Spec.Template = newDeployment.Spec.Template
		if err := r.Update(ctx, currentDeployment); err != nil {
			r.Log.Error(err, "Failed to update Deployment", "Deployment.Namespace", currentDeployment.Namespace, "Deployment.Name", currentDeployment.Name)
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

func getWorkingSorucesDir(r serverlessv1alpha2.Runtime) string {
	switch r {
	case serverlessv1alpha2.NodeJs20:
		return "/usr/src/app/function"
	default:
		return ""
	}
}

func getFunctionSource(r serverlessv1alpha2.Runtime) string {
	switch r {
	case serverlessv1alpha2.NodeJs20:
		return `
const _ = require('lodash')
	module.exports = {
	main: function(event, context) {
			return _.kebabCase('Hello World from Node.js 20 Function');
		}
	}`
	default:
		return ""
	}
}

func getFunctionDependencies(r serverlessv1alpha2.Runtime) string {
	switch r {
	case serverlessv1alpha2.NodeJs20:
		return `
{
  "name": "test-function-nodejs",
  "version": "1.0.0",
  "dependencies": {
	"lodash":"^4.17.20"
  }
}`
	default:
		return ""
	}
}

func getRuntimeCommand(r serverlessv1alpha2.Runtime) string {
	switch r {
	case serverlessv1alpha2.NodeJs20:
		return `
printf "${FUNCTION_SOURCE}" > handler.js;
printf "${FUNCTION_DEPENDENCIES}" > package.json;
npm install --prefer-offline --no-audit --progress=false;
cd ..;
npm start;
`
	default:
		return ""
	}
}

func (r *FunctionReconciler) constructDeploymentForFunction(function *serverlessv1alpha2.Function) *appsv1.Deployment {
	fRuntime := function.Spec.Runtime

	workingSorucesDir := getWorkingSorucesDir(fRuntime)

	functionDependencies := getFunctionDependencies(fRuntime)

	functionSource := getFunctionSource(fRuntime)

	runtimeCommand := getRuntimeCommand(fRuntime)

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
					Volumes: []v1.Volume{
						{
							// used for wiriting sources (code&deps) to the sources dir
							Name: "sources",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						},
					},
					Containers: []v1.Container{
						{
							Name:  fmt.Sprintf("%s-function-pod", function.Name),
							Image: "europe-docker.pkg.dev/kyma-project/prod/function-runtime-nodejs20:main",
							//Env:        r.getRuntimeEnvs(f),
							WorkingDir: workingSorucesDir,
							Command: []string{
								"sh",
								"-c",
								`
printf "${FUNCTION_SOURCE}" > handler.js;
printf "${FUNCTION_DEPENDENCIES}" > package.json;
npm install --prefer-offline --no-audit --progress=false;
cd ..;
npm start;
`,
							},
							Env: []v1.EnvVar{
								{
									Name:  "FUNCTION_SOURCE",
									Value: functionSource,
								},
								{
									Name:  "FUNCTION_DEPENDENCIES",
									Value: functionDependencies,
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "sources",
									MountPath: workingSorucesDir,
								},
							},
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
