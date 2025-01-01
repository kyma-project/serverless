package resources

import (
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/internal/config"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"path"
)

const DefaultDeploymentReplicas int32 = 1

type Deployment struct {
	*appsv1.Deployment
	functionConfig *config.FunctionConfig
	function       *serverlessv1alpha2.Function
}

func NewDeployment(f *serverlessv1alpha2.Function, c *config.FunctionConfig) *Deployment {
	d := &Deployment{
		functionConfig: c,
		function:       f,
	}
	d.Deployment = d.construct()
	return d
}

func (d *Deployment) construct() *appsv1.Deployment {
	labels := map[string]string{
		"app": d.name(),
		// TODO: do we need to add more labels here?
		serverlessv1alpha2.FunctionNameLabel:      d.function.GetName(),
		serverlessv1alpha2.FunctionManagedByLabel: serverlessv1alpha2.FunctionControllerValue,
		serverlessv1alpha2.FunctionResourceLabel:  "",
		serverlessv1alpha2.FunctionUUIDLabel:      "",
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.name(),
			Namespace: d.function.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: d.podSpec(),
			},
			Replicas: d.replicas(),
		},
	}
	return deployment
}

func (d *Deployment) RuntimeImage() string {
	return d.Spec.Template.Spec.Containers[0].Image
}

func (d *Deployment) name() string {
	return d.function.Name
}

func (d *Deployment) podSpec() corev1.PodSpec {
	secretVolumes, secretVolumeMounts := d.deploymentSecretVolumes()

	return corev1.PodSpec{
		Volumes: append(d.volumes(), secretVolumes...),
		Containers: []corev1.Container{
			{
				Name:       d.name(),
				Image:      d.runtimeImage(),
				WorkingDir: d.workingSourcesDir(),
				Command: []string{
					"sh",
					"-c",
					d.runtimeCommand(),
				},
				Resources:    d.resourceConfiguration(),
				Env:          d.envs(),
				VolumeMounts: append(d.volumeMounts(), secretVolumeMounts...),
				Ports: []corev1.ContainerPort{
					{
						ContainerPort: 80,
					},
				},
				//TODO: add SecurityContext
			},
		},
	}
}

func (d *Deployment) replicas() *int32 {
	replicas := d.function.Spec.Replicas
	if replicas != nil {
		return replicas
	}
	defaultValue := DefaultDeploymentReplicas
	return &defaultValue
}

func (d *Deployment) volumes() []corev1.Volume {
	runtime := d.function.Spec.Runtime
	volumes := []corev1.Volume{
		{
			// used for writing sources (code&deps) to the sources dir
			Name: "sources",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "package-registry-config",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: d.functionConfig.PackageRegistryConfigSecretName,
					Optional:   ptr.To[bool](true),
				},
			},
		},
	}
	if runtime == serverlessv1alpha2.Python312 {
		volumes = append(volumes, corev1.Volume{
			// required by pip to save deps to .local dir
			Name: "local",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}
	return volumes
}

func (d *Deployment) volumeMounts() []corev1.VolumeMount {
	runtime := d.function.Spec.Runtime
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "sources",
			MountPath: d.workingSourcesDir(),
		},
	}
	if runtime == serverlessv1alpha2.NodeJs20 {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "package-registry-config",
			ReadOnly:  true,
			MountPath: path.Join(d.workingSourcesDir(), "package-registry-config/.npmrc"),
			SubPath:   ".npmrc",
		})
	}
	if runtime == serverlessv1alpha2.Python312 {
		volumeMounts = append(volumeMounts,
			corev1.VolumeMount{
				Name:      "local",
				MountPath: "/.local",
			},
			corev1.VolumeMount{
				Name:      "package-registry-config",
				ReadOnly:  true,
				MountPath: path.Join(d.workingSourcesDir(), "package-registry-config/pip.conf"),
				SubPath:   "pip.conf",
			})
	}
	return volumeMounts
}

func (d *Deployment) runtimeImage() string {
	runtimeOverride := d.function.Spec.RuntimeImageOverride
	if runtimeOverride != "" {
		return runtimeOverride
	}

	switch d.function.Spec.Runtime {
	case serverlessv1alpha2.NodeJs20:
		return d.functionConfig.ImageNodeJs20
	case serverlessv1alpha2.Python312:
		return d.functionConfig.ImagePython312
	default:
		return ""
	}
}

func (d *Deployment) workingSourcesDir() string {
	switch d.function.Spec.Runtime {
	case serverlessv1alpha2.NodeJs20:
		return "/usr/src/app/function"
	case serverlessv1alpha2.Python312:
		return "/kubeless"
	default:
		return ""
	}
}

func (d *Deployment) runtimeCommand() string {
	spec := &d.function.Spec
	dependencies := spec.Source.Inline.Dependencies
	switch spec.Runtime {
	case serverlessv1alpha2.NodeJs20:
		if dependencies != "" {
			// if deps are not empty use pip
			return `printf "${FUNC_HANDLER_SOURCE}" > handler.js;
printf "${FUNC_HANDLER_DEPENDENCIES}" > package.json;
npm install --prefer-offline --no-audit --progress=false;
cd ..;
npm start;`
		}
		return `printf "${FUNC_HANDLER_SOURCE}" > handler.js;
cd ..;
npm start;`
	case serverlessv1alpha2.Python312:
		if dependencies != "" {
			// if deps are not empty use npm
			return `printf "${FUNC_HANDLER_SOURCE}" > handler.py;
printf "${FUNC_HANDLER_DEPENDENCIES}" > requirements.txt;
PIP_CONFIG_FILE=package-registry-config/pip.conf pip install --user --no-cache-dir -r /kubeless/requirements.txt;
cd ..;
python /kubeless.py;`
		}
		return `printf "${FUNC_HANDLER_SOURCE}" > handler.py;
cd ..;
python /kubeless.py;`
	default:
		return ""
	}
}

func (d *Deployment) envs() []corev1.EnvVar {
	spec := &d.function.Spec
	envs := []corev1.EnvVar{
		{
			Name:  "FUNC_HANDLER_SOURCE",
			Value: spec.Source.Inline.Source,
		},
		{
			Name:  "FUNC_HANDLER_DEPENDENCIES",
			Value: spec.Source.Inline.Dependencies,
		},
	}
	if spec.Runtime == serverlessv1alpha2.Python312 {
		envs = append(envs, []corev1.EnvVar{
			{
				Name:  "MOD_NAME",
				Value: "handler",
			},
			{
				Name:  "FUNC_HANDLER",
				Value: "main",
			},
		}...)
	}
	envs = append(envs, spec.Env...)
	return envs
}

func (d *Deployment) resourceConfiguration() corev1.ResourceRequirements {
	resCfg := d.function.Spec.ResourceConfiguration
	if resCfg != nil && resCfg.Function != nil && resCfg.Function.Resources != nil {
		return *resCfg.Function.Resources
	}
	return corev1.ResourceRequirements{}
}

func (d *Deployment) deploymentSecretVolumes() (volumes []corev1.Volume, volumeMounts []corev1.VolumeMount) {
	volumes = []corev1.Volume{}
	volumeMounts = []corev1.VolumeMount{}
	for _, secretMount := range d.function.Spec.SecretMounts {
		volumeName := secretMount.SecretName

		volume := corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  secretMount.SecretName,
					DefaultMode: ptr.To[int32](0666), //read and write only for everybody
					Optional:    ptr.To[bool](false),
				},
			},
		}
		volumes = append(volumes, volume)

		volumeMount := corev1.VolumeMount{
			Name:      volumeName,
			ReadOnly:  true,
			MountPath: secretMount.MountPath,
		}
		volumeMounts = append(volumeMounts, volumeMount)
	}
	return volumes, volumeMounts
}
