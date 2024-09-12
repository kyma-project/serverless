package controllers

import (
	"context"
	"fmt"
	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/pkg/apis/serverless/v1alpha2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"os"
	"path"
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

var (
	svcTargetPort = intstr.FromInt32(8080)
)

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

	if !instanceHasFinalizer {
		log.Info("adding finalizer")
		controllerutil.AddFinalizer(&f, finalizerName)
		errUpdF := r.client.Update(ctx, &f)
		if errUpdF != nil {
			return reconcile.Result{}, errors.Wrap(errUpdF, "unable to update Function")
		}
	}

	log.Info("starting patching PersistentVolume")
	errCreatePV := r.createPV(ctx, f)
	if errCreatePV != nil {
		return reconcile.Result{}, errCreatePV
	}

	log.Info("starting patching PersistentVolumeClaim")
	errCreatePVC := r.createPVC(ctx, f)
	if errCreatePVC != nil {
		return reconcile.Result{}, errCreatePVC
	}

	log.Info("starting writing sources to NFS")
	result, err, done := r.writeSourceToPVC(log, f)
	if done {
		return result, err
	}

	log.Info("starting creating Deployment")
	errCreateDeployment := r.createDeployment(ctx, f)
	if errCreateDeployment != nil {
		return reconcile.Result{}, errCreateDeployment
	}

	log.Info("starting creating Service")
	errCreateService := r.createService(ctx, f)
	if errCreateService != nil {
		return reconcile.Result{}, errCreateService
	}

	// TODO: create
	// 4.1 check if deployment is ready
	// 5. update status

	// TODO: update
	// 3.1 check if deployment is ready
	// 4. update status

	log.Info("end reconciliation")
	return reconcile.Result{}, nil
}

func (r *FunctionReconciler) createService(ctx context.Context, f serverlessv1alpha2.Function) error {
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      f.GetName(),
			Namespace: f.GetNamespace(),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       "http", // it has to be here for istio to work properly
				TargetPort: svcTargetPort,
				Port:       80,
				Protocol:   corev1.ProtocolTCP,
			}},
			Selector: map[string]string{"function": f.GetName()},
		},
	}
	errPatchService := r.client.Patch(ctx, service, client.Apply, &client.PatchOptions{
		Force:        ptr.To(true),
		FieldManager: "buildless-serverless-controller",
	})
	if errPatchService != nil {
		return errors.Wrap(errPatchService, "unable to patch Service")
	}
	return nil
}

func (r *FunctionReconciler) createDeployment(ctx context.Context, f serverlessv1alpha2.Function) error {
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
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      f.GetName(),
			Namespace: f.GetNamespace(),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"function": f.GetName()},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"function": f.GetName()},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "sources",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: f.GetName(),
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "function-container",
							Image: "europe-docker.pkg.dev/kyma-project/prod/function-runtime-nodejs20:main",
							Env:   envs,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "sources",
									MountPath: "/usr/src/app/function",
								},
							},
							Command: []string{
								"sh",
								"-c",
								"cd /usr/src/app/function; npm install; cd ..; npm start",
							},
							/*
								In order to mark pod as ready we need to ensure the function is actually running and ready to serve traffic.
								We do this but first ensuring that sidecar is ready by using "proxy.istio.io/config": "{ \"holdApplicationUntilProxyStarts\": true }", annotation
								Second thing is setting that probe which continuously
							*/
							StartupProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/healthz",
										Port: svcTargetPort,
									},
								},
								InitialDelaySeconds: 0,
								PeriodSeconds:       5,
								SuccessThreshold:    1,
								FailureThreshold:    30, // FailureThreshold * PeriodSeconds = 150s in this case, this should be enough for any function pod to start up
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/healthz",
										Port: svcTargetPort,
									},
								},
								InitialDelaySeconds: 0, // startup probe exists, so delaying anything here doesn't make sense
								FailureThreshold:    1,
								PeriodSeconds:       5,
								TimeoutSeconds:      2,
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/healthz",
										Port: svcTargetPort,
									},
								},
								FailureThreshold: 3,
								PeriodSeconds:    5,
								TimeoutSeconds:   4,
							},
							ImagePullPolicy: corev1.PullIfNotPresent,
						},
					},
				},
			},
		},
	}
	errPatchDeployment := r.client.Patch(ctx, deployment, client.Apply, &client.PatchOptions{
		Force:        ptr.To(true),
		FieldManager: "buildless-serverless-controller",
	})
	if errPatchDeployment != nil {
		return errors.Wrap(errPatchDeployment, "unable to patch Deployment")
	}
	return nil
}

func (r *FunctionReconciler) createPVC(ctx context.Context, f serverlessv1alpha2.Function) error {
	pvc := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      f.GetName(),
			Namespace: f.GetNamespace(),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{"storage": resource.MustParse("5Gi")},
			},
			VolumeName:       f.GetName(),
			StorageClassName: ptr.To(""),
		},
	}
	errPatchPVC := r.client.Patch(ctx, pvc, client.Apply, &client.PatchOptions{
		Force:        ptr.To(true),
		FieldManager: "buildless-serverless-controller",
	})
	if errPatchPVC != nil {
		return errors.Wrap(errPatchPVC, "unable to patch PersistentVolumeClaim")
	}
	return nil
}

func (r *FunctionReconciler) createPV(ctx context.Context, f serverlessv1alpha2.Function) error {
	pv := &corev1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolume",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      f.GetName(),
			Namespace: f.GetNamespace(),
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: corev1.ResourceList{"storage": resource.MustParse("5Gi")},
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				NFS: &corev1.NFSVolumeSource{
					Server: r.nfsServiceIP, // requires svc ip ( name.namespace.svc.cluster.local is not supported )
					Path:   fmt.Sprintf("/%s", f.GetUID()),
				},
			},
			AccessModes:  []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
			MountOptions: []string{"nfsvers=4.2"}, // required!!
		},
	}
	errPatchPV := r.client.Patch(ctx, pv, client.Apply, &client.PatchOptions{
		Force:        ptr.To(true),
		FieldManager: "buildless-serverless-controller",
	})
	if errPatchPV != nil {
		return errors.Wrap(errPatchPV, "unable to patch PersistentVolume")
	}
	return nil
}

func (r *FunctionReconciler) writeSourceToPVC(log *logrus.Entry, f serverlessv1alpha2.Function) (reconcile.Result, error, bool) {
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
	errWritePackages := os.WriteFile(path.Join(functionSourcePath, "package.json"), []byte(f.Spec.Source.Inline.Dependencies), os.ModePerm)
	if errWritePackages != nil {
		return reconcile.Result{}, errors.Wrap(errWritePackages, "unable to write package.json"), true
	}

	log.Info("end writing sources to ", functionSourcePath)
	return reconcile.Result{}, nil, false
}
