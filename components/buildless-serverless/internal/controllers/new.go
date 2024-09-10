package controllers

import (
	"context"

	"github.com/kyma-project/serverless/components/buildless-serverless/pkg/apis/serverless/v1alpha2"
	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/pkg/apis/serverless/v1alpha2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type FunctionReconciler struct {
	client     client.Client
	initLogger *logrus.Logger
	cfg        ReconcilerConfig
}

var (
	svcTargetPort = intstr.FromInt32(8080)
)

type ReconcilerConfig struct {
	Nodejs20Image  string
	Python312Image string
}

func NewFunctionReconciler(client client.Client, logger *logrus.Logger, cfg ReconcilerConfig) *FunctionReconciler {
	return &FunctionReconciler{
		client:     client,
		initLogger: logger,
		cfg:        cfg,
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

	if instanceIsBeingDeleted {
		return reconcile.Result{}, nil
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
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: f.APIVersion,
					Kind:       f.Kind,
					Name:       f.GetName(),
					UID:        f.GetUID(),
				},
			},
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

func (r *FunctionReconciler) getRuntimeCommand(f v1alpha2.Function) string {
	if f.Spec.Runtime == v1alpha2.NodeJs20 {
		if f.Spec.Source.Inline.Dependencies != "" {
			// if deps are not empty use pip
			return `
printf "${FUNC_HANDLER_SOURCE}" > handler.js;
printf "${FUNC_HANDLER_DEPENDENCIES}" > package.json;
npm install --prefer-offline --no-audit --progress=false;
cd ..;
npm start;
`
		}

		return `
printf "${FUNC_HANDLER_SOURCE}" > handler.js;
cd ..;
npm start;
`
	}

	if f.Spec.Source.Inline.Dependencies != "" {
		// if deps are not empty use npm
		return `
printf "${FUNC_HANDLER_SOURCE}" > handler.py;
printf "${FUNC_HANDLER_DEPENDENCIES}" > requirements.txt;
pip install --user --no-cache-dir -r /kubeless/requirements.txt;
cd ..;
python /kubeless.py;
`
	}

	return `
printf "${FUNC_HANDLER_SOURCE}" > handler.py;
cd ..;
python /kubeless.py;
`
}

func (r *FunctionReconciler) getRuntimeImage(runtime v1alpha2.Runtime) string {
	if runtime == v1alpha2.NodeJs20 {
		return r.cfg.Nodejs20Image
	}

	return r.cfg.Python312Image
}

func (r *FunctionReconciler) getRuntimeSourcesDir(runtime v1alpha2.Runtime) string {
	if runtime == v1alpha2.NodeJs20 {
		return "/usr/src/app/function"
	}

	return "/kubeless"
}

func (r *FunctionReconciler) getRuntimeEnvs(f v1alpha2.Function) []corev1.EnvVar {
	if f.Spec.Runtime == v1alpha2.NodeJs20 {
		deps := f.Spec.Source.Inline.Dependencies
		// npm has problems when deps are empty or not in JSON format
		if deps == "" {
			deps = "{}\n"
		}

		return []corev1.EnvVar{
			{Name: "FUNC_RUNTIME", Value: string(f.Spec.Runtime)},
			{Name: "FUNC_NAME", Value: f.Name},
			{Name: "SERVICE_NAMESPACE", Value: f.Namespace},
			{Name: "TRACE_COLLECTOR_ENDPOINT", Value: ""},
			{Name: "PUBLISHER_PROXY_ADDRESS", Value: ""},
			{Name: "FUNC_HANDLER", Value: "main"},
			{Name: "MOD_NAME", Value: "handler"},
			{Name: "FUNC_PORT", Value: "8080"},
			{Name: "FUNC_HANDLER_SOURCE", Value: f.Spec.Source.Inline.Source},
			{Name: "FUNC_HANDLER_DEPENDENCIES", Value: deps},
		}
	}

	return []corev1.EnvVar{
		{Name: "FUNC_RUNTIME", Value: string(f.Spec.Runtime)},
		{Name: "PYTHONPATH", Value: "$(KUBELESS_INSTALL_VOLUME)/lib.python3.12/site-packages:$(KUBELESS_INSTALL_VOLUME)"},
		{Name: "PYTHONUNBUFFERED", Value: "TRUE"},
		{Name: "FUNC_NAME", Value: f.Name},
		{Name: "SERVICE_NAMESPACE", Value: f.Namespace},
		{Name: "TRACE_COLLECTOR_ENDPOINT", Value: ""},
		{Name: "PUBLISHER_PROXY_ADDRESS", Value: ""},
		{Name: "FUNC_HANDLER", Value: "main"},
		{Name: "MOD_NAME", Value: "handler"},
		{Name: "FUNC_PORT", Value: "8080"},
		{Name: "FUNC_HANDLER_SOURCE", Value: f.Spec.Source.Inline.Source},
		{Name: "FUNC_HANDLER_DEPENDENCIES", Value: f.Spec.Source.Inline.Dependencies},
	}
}

func (r *FunctionReconciler) createDeployment(ctx context.Context, f serverlessv1alpha2.Function) error {
	resources := &corev1.ResourceRequirements{}
	if f.Spec.ResourceConfiguration != nil &&
		f.Spec.ResourceConfiguration.Function != nil &&
		f.Spec.ResourceConfiguration.Function.Resources != nil {
		resources = f.Spec.ResourceConfiguration.Function.Resources
	}

	workingSourcesDir := r.getRuntimeSourcesDir(f.Spec.Runtime)

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      f.GetName(),
			Namespace: f.GetNamespace(),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: f.APIVersion,
					Kind:       f.Kind,
					Name:       f.GetName(),
					UID:        f.GetUID(),
				},
			},
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
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:  ptr.To(int64(1001)),
						RunAsGroup: ptr.To(int64(1001)),
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
					Volumes: []corev1.Volume{
						{
							// used for wiriting sources (code&deps) to the sources dir
							Name: "sources",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							// required by pip to save deps to .local dir
							Name: "local",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							// required by python
							Name: "tmp-dir",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									SizeLimit: resource.NewScaledQuantity(100, resource.Mega),
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:       "function-container",
							Image:      r.getRuntimeImage(f.Spec.Runtime),
							Env:        r.getRuntimeEnvs(f),
							WorkingDir: workingSourcesDir,
							Command: []string{
								"sh",
								"-c",
								r.getRuntimeCommand(f),
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "sources",
									MountPath: workingSourcesDir,
								},
								{
									Name:      "local",
									MountPath: "/.local",
								},
								{
									Name:      "tmp-dir",
									MountPath: "/tmp",
								},
							},
							SecurityContext: &corev1.SecurityContext{
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{
										"ALL",
									},
								},
								Privileged:             ptr.To(false),
								ProcMount:              ptr.To(corev1.DefaultProcMount),
								ReadOnlyRootFilesystem: ptr.To(true),
							},
							Resources: *resources,
							/*
								I changed Probes configuration to easly mesure performance
								TODO: adjust probes configuration
							*/
							StartupProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/healthz",
										Port: svcTargetPort,
									},
								},
								InitialDelaySeconds: 0,
								PeriodSeconds:       1,
								SuccessThreshold:    1,
								FailureThreshold:    300, // FailureThreshold * PeriodSeconds = 300s in this case, this should be enough for any function pod to start up
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
