package state

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

type deploymentBuilder struct {
	functionConfig config.FunctionConfig
	instance       *serverlessv1alpha2.Function
}

func NewDeploymentBuilder(m *stateMachine) *deploymentBuilder {
	return &deploymentBuilder{
		functionConfig: m.functionConfig,
		instance:       &m.state.instance,
	}
}

func (b *deploymentBuilder) build() *appsv1.Deployment {
	labels := map[string]string{
		"app": b.deploymentName(),
		// TODO: do we need to add more labels here?
		serverlessv1alpha2.FunctionNameLabel: b.instance.GetName(),
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.deploymentName(),
			Namespace: b.instance.Namespace,
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
				Spec: b.buildPodSpec(),
			},
			Replicas: b.getReplicas(),
		},
	}
	return deployment
}

func (b *deploymentBuilder) deploymentName() string {
	return b.instance.Name
}

func (b *deploymentBuilder) buildPodSpec() corev1.PodSpec {
	secretVolumes, secretVolumeMounts := b.buildDeploymentSecretVolumes()

	return corev1.PodSpec{
		Volumes: append(b.getVolumes(), secretVolumes...),
		Containers: []corev1.Container{
			{
				Name:       b.deploymentName(),
				Image:      b.getRuntimeImage(),
				WorkingDir: b.getWorkingSourcesDir(),
				Command: []string{
					"sh",
					"-c",
					b.getRuntimeCommand(),
				},
				Resources:    b.getResourceConfiguration(),
				Env:          b.getEnvs(),
				VolumeMounts: append(b.getVolumeMounts(), secretVolumeMounts...),
				Ports: []corev1.ContainerPort{
					{
						ContainerPort: 80,
					},
				},
				//TODO: uncomment later - now we need greater privileges for running npm command
				// SecurityContext: b.restrictiveContainerSecurityContext(),
			},
		},
	}
}

func (b *deploymentBuilder) getReplicas() *int32 {
	replicas := &b.instance.Spec.Replicas
	if replicas != nil {
		return *replicas
	}
	defaultValue := DefaultDeploymentReplicas
	return &defaultValue
}

func (b *deploymentBuilder) getVolumes() []corev1.Volume {
	runtime := b.instance.Spec.Runtime
	volumes := []corev1.Volume{
		{
			// used for writing sources (code&deps) to the sources dir
			Name: "sources",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "registry-config",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: b.functionConfig.PackageRegistryConfigSecretName,
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

func (b *deploymentBuilder) getVolumeMounts() []corev1.VolumeMount {
	runtime := b.instance.Spec.Runtime
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "sources",
			MountPath: b.getWorkingSourcesDir(),
		},
	}
	if runtime == serverlessv1alpha2.NodeJs20 {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "registry-config",
			ReadOnly:  true,
			MountPath: path.Join(b.getWorkingSourcesDir(), "registry-config/.npmrc"),
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
				Name:      "registry-config",
				ReadOnly:  true,
				MountPath: path.Join(b.getWorkingSourcesDir(), "registry-config/pip.conf"),
				SubPath:   "pip.conf",
			})
	}
	return volumeMounts
}

func (b *deploymentBuilder) getRuntimeImage() string {
	runtimeOverride := b.instance.Spec.RuntimeImageOverride
	if runtimeOverride != "" {
		return runtimeOverride
	}

	switch b.instance.Spec.Runtime {
	case serverlessv1alpha2.NodeJs20:
		return b.functionConfig.ImageNodeJs20
	case serverlessv1alpha2.Python312:
		return b.functionConfig.ImagePython312
	default:
		return ""
	}
}

func (b *deploymentBuilder) getWorkingSourcesDir() string {
	switch b.instance.Spec.Runtime {
	case serverlessv1alpha2.NodeJs20:
		return "/usr/src/app/function"
	case serverlessv1alpha2.Python312:
		return "/kubeless"
	default:
		return ""
	}
}

func (b *deploymentBuilder) getRuntimeCommand() string {
	spec := &b.instance.Spec
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
cp registry-config/pip.conf .;
pip install --user --no-cache-dir -r /kubeless/requirements.txt;
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

func (b *deploymentBuilder) getEnvs() []corev1.EnvVar {
	spec := &b.instance.Spec
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

func (b *deploymentBuilder) getResourceConfiguration() corev1.ResourceRequirements {
	resCfg := b.instance.Spec.ResourceConfiguration
	if resCfg != nil && resCfg.Function != nil && resCfg.Function.Resources != nil {
		return *resCfg.Function.Resources
	}
	return corev1.ResourceRequirements{}
}

func (b *deploymentBuilder) buildDeploymentSecretVolumes() (volumes []corev1.Volume, volumeMounts []corev1.VolumeMount) {
	volumes = []corev1.Volume{}
	volumeMounts = []corev1.VolumeMount{}
	for _, secretMount := range b.instance.Spec.SecretMounts {
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

// security context is set to fulfill the baseline security profile
// based on https://raw.githubusercontent.com/kyma-project/community/main/concepts/psp-replacement/baseline-pod-spec.yaml
func (b *deploymentBuilder) restrictiveContainerSecurityContext() *corev1.SecurityContext {
	defaultProcMount := corev1.DefaultProcMount
	return &corev1.SecurityContext{
		Privileged: ptr.To[bool](false),
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{
				"ALL",
			},
		},
		ProcMount:              &defaultProcMount,
		ReadOnlyRootFilesystem: ptr.To[bool](true),
	}
}
