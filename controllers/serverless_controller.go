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
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	"github.com/kyma-project/module-manager/pkg/declarative"
	rtypes "github.com/kyma-project/module-manager/pkg/types"
	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/prerequisites"
)

const (
	chartNs         = "kyma-system"
	requeueInterval = time.Second * 3
)

// ServerlessReconciler reconciles a Serverless object
type ServerlessReconciler struct {
	declarative.ManifestReconciler
	client.Client
	Scheme *runtime.Scheme
	*rest.Config
	ChartPath string
}

// TODO: serverless-manager doesn't need almost half of these rbscs. It uses them only to create another rbacs ( is there any onther option? - investigate )

//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="",resources=services;secrets;serviceaccounts;configmaps,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete;deletecollection

//+kubebuilder:rbac:groups=apps,resources=replicasets,verbs=list
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete;deletecollection

//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete;deletecollection

//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get

//+kubebuilder:rbac:groups=policy,resources=podsecuritypolicies,verbs=use

//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings;roles,verbs=get;list;watch;create;update;patch;delete;deletecollection

//+kubebuilder:rbac:groups=serverless.kyma-project.io,resources=functions,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=serverless.kyma-project.io,resources=functions/status,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=serverless.kyma-project.io,resources=gitrepositories,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=serverless.kyma-project.io,resources=gitrepositories/status,verbs=get

//+kubebuilder:rbac:groups=networking.istio.io,resources=destinationrules,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices,verbs=get;list;watch;create;update;patch;delete;deletecollection

//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=serverlesses,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=serverlesses/status,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=serverlesses/finalizers,verbs=get;list;watch;create;update;patch;delete;deletecollection

//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations;mutatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete;deletecollection

//+kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch;create;update;patch;delete;deletecollection

//+kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch;delete;deletecollection

//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;update;patch;delete;deletecollection

func (r *ServerlessReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var serverless v1alpha1.Serverless
	namespacedName := types.NamespacedName{Namespace: req.Namespace, Name: req.Name}
	err := r.Client.Get(ctx, namespacedName, &serverless)
	if err != nil {
		return ctrl.Result{RequeueAfter: requeueInterval}, err
	}

	err = checkPrerequisites(ctx, r.Client, &serverless)
	if err != nil {
		return ctrl.Result{RequeueAfter: requeueInterval}, err
	}

	return r.ManifestReconciler.Reconcile(ctx, req)
}

func checkPrerequisites(ctx context.Context, client client.Client, serverless *v1alpha1.Serverless) error {
	if serverless.Status.State != rtypes.StateProcessing &&
		serverless.Status.State != "" {
		return nil
	}

	return prerequisites.Check(ctx, client, serverless)
}

// initReconciler injects the required configuration into the declarative reconciler.
func (r *ServerlessReconciler) initReconciler(mgr ctrl.Manager) error {
	manifestResolver := &ManifestResolver{
		chartPath: r.ChartPath,
	}

	return r.Inject(mgr, &v1alpha1.Serverless{},
		declarative.WithManifestResolver(manifestResolver),
		declarative.WithResourcesReady(true),
		declarative.WithFinalizer("serverless-manager.kyma-project.io/deletion-hook"),
	)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServerlessReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Config = mgr.GetConfig()
	if err := r.initReconciler(mgr); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Serverless{}).
		Complete(r)
}

// ManifestResolver represents the chart information for the passed Sample resource.
type ManifestResolver struct {
	chartPath string
}

// Get returns the chart information to be processed.
func (m *ManifestResolver) Get(obj rtypes.BaseCustomObject, _ logr.Logger) (rtypes.InstallationSpec, error) {
	serverless, valid := obj.(*v1alpha1.Serverless)
	if !valid {
		return rtypes.InstallationSpec{},
			fmt.Errorf("invalid type conversion for %s", client.ObjectKeyFromObject(obj))
	}

	// default empty fields
	serverless.Spec.Default()

	flags, err := structToFlags(serverless.Spec)
	if err != nil {
		return rtypes.InstallationSpec{},
			fmt.Errorf("resolving manifest failed: %w", err)
	}

	return rtypes.InstallationSpec{
		ChartPath: m.chartPath,
		ChartFlags: rtypes.ChartFlags{
			ConfigFlags: rtypes.Flags{
				"Namespace":       chartNs,
				"CreateNamespace": false, // TODO: think about it
			},
			SetFlags: flags,
		},
	}, nil
}

func structToFlags(obj interface{}) (flags rtypes.Flags, err error) {
	data, err := json.Marshal(obj)

	if err != nil {
		return
	}

	err = json.Unmarshal(data, &flags)
	return
}
