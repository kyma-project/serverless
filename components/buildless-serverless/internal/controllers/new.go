package controllers

import (
	"context"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/pkg/apis/serverless/v1alpha2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type FunctionReconciler struct {
	client              client.Client
	initLogger          *logrus.Logger
	functionSourcesPath string
	nfsServiceIP        string
}

const finalizerName = "serverless.kyma-project.io/function-finalizer"

func NewFunctionReconciler(client client.Client, logger *logrus.Logger, functionSourcesPath string, nfsServiceIP string) *FunctionReconciler {
	return &FunctionReconciler{
		client:              client,
		initLogger:          logger,
		functionSourcesPath: functionSourcesPath,
		nfsServiceIP:        nfsServiceIP,
	}
}

func (r *FunctionReconciler) SetupWithManager(mgr manager.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("function-controller").
		For(&serverlessv1alpha2.Function{}, builder.WithPredicates(predicate.Funcs{UpdateFunc: IsNotFunctionStatusUpdate(r.initLogger)})).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

func (r *FunctionReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log := r.initLogger.WithField("function", request.NamespacedName)
	log.Info("starting reconciliation")
	result, err := r.reconcile(ctx, log, request)
	log.Info("end reconciliation")
	return result, err
}

func (r *FunctionReconciler) reconcile(ctx context.Context, log *logrus.Entry, request reconcile.Request) (reconcile.Result, error) {
	var f serverlessv1alpha2.Function
	errGetF := r.client.Get(ctx, request.NamespacedName, &f)
	if errGetF != nil {
		return reconcile.Result{}, errors.Wrap(errGetF, "unable to fetch Function")
	}

	instanceIsBeingDeleted := !f.GetDeletionTimestamp().IsZero()
	instanceHasFinalizer := controllerutil.ContainsFinalizer(&f, finalizerName)

	if instanceIsBeingDeleted {
		if !instanceHasFinalizer {
			return reconcile.Result{}, nil
		}

		// TODO: delete

		// 1. remove deploy & service
		// 2. cleanup pvc
		// 3. remove finalizer

		return reconcile.Result{}, nil
	}

	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nfs",
			Namespace: f.GetNamespace(),
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: corev1.ResourceList{"storage": resource.MustParse("5Gi")},
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				NFS: &corev1.NFSVolumeSource{
					Server: r.nfsServiceIP, // requires svc ip ( name.namespace.svc.cluster.local is not supported )
					Path:   "/",
				},
			},
			AccessModes:                   []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
			MountOptions:                  []string{"nfsvers=4.2"}, // required!!
		},
	}
	errr.client.Create(ctx, pv)


	pvc :=
	//apiVersion: v1
	//kind: PersistentVolumeClaim
	//metadata:
	//  name: nfs
	//spec:
	//  accessModes:
	//    - ReadWriteMany
	//  storageClassName: ""
	//  resources:
	//    requests:
	//      storage: 5Gi
	//  volumeName: nfs

	if !instanceHasFinalizer {
		log.Info("adding finalizer")
		controllerutil.AddFinalizer(&f, finalizerName)
		errUpdF := r.client.Update(ctx, &f)
		if errUpdF != nil {
			return reconcile.Result{}, errors.Wrap(errUpdF, "unable to update Function")
		}
	}

	result, err, done := r.writeSourceToPvc(log, f)
	if done {
		return result, err
	}

	envs := []corev1.EnvVar{
		{Name: "FUNC_RUNTIME", Value: string(f.Spec.Runtime)},
		{Name: "FUNC_NAME", Value: f.Name},
		{Name: "SERVICE_NAMESPACE", Value: f.Namespace},
		{Name: "TRACE_COLLECTOR_ENDPOINT", Value: ""},
		{Name: "PUBLISHER_PROXY_ADDRESS", Value: ""},
		{Name: "FUNC_HANDLER", Value: "main"},
		{Name: "MOD_NAME", Value: "handler"},
		{Name: "FUNC_PORT", Value: "8080"},
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       appsv1.DeploymentSpec{},
	}

	// TODO: create
	// 3.2. ??? run npm install
	// 4. create deploy & service
	// 4.1 check if deployment is ready
	// 5. update status

	// TODO: update
	// 2.2. ??? rerun npm install
	// 3. ??? update deploy ( we should have hot-reload enabled on every runtime )
	// 3.1 check if deployment is ready
	// 4. update status

	return reconcile.Result{}, nil
}

func (r *FunctionReconciler) writeSourceToPvc(log *logrus.Entry, f serverlessv1alpha2.Function) (reconcile.Result, error, bool) {
	functionSourcePath := path.Join(r.functionSourcesPath, string(f.GetUID()))
	log.Info("starting writing sources to ", functionSourcePath)

	errMkdir := os.Mkdir(functionSourcePath, os.ModePerm)
	if errMkdir != nil && !os.IsExist(errMkdir) {
		return reconcile.Result{}, errors.Wrap(errMkdir, "unable to create directory for Function source"), true
	}
	// TODO: add support for git functions
	// TODO: add support for python
	errWriteHandler := os.WriteFile(path.Join(functionSourcePath, "handler.js"), []byte(f.Spec.Source.Inline.Source), os.ModePerm)
	if errWriteHandler != nil {
		return reconcile.Result{}, errors.Wrap(errWriteHandler, "unable to write handler.js"), true
	}
	errWritePackages := os.WriteFile(path.Join(functionSourcePath, "package.js"), []byte(f.Spec.Source.Inline.Dependencies), os.ModePerm)
	if errWritePackages != nil {
		return reconcile.Result{}, errors.Wrap(errWritePackages, "unable to write package.js"), true
	}

	log.Info("end writing sources to ", functionSourcePath)
	return reconcile.Result{}, nil, false
}
