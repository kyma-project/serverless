package state

import (
	"fmt"
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

const DefaultDeploymentReplicas int32 = 1

func (m *stateMachine) buildDeployment() *appsv1.Deployment {
	f := &m.state.instance
	labels := map[string]string{
		"app": f.Name,
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-function-deployment", f.Name),
			Namespace: f.Namespace,
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
				Spec: m.buildPodSpec(),
			},
			Replicas: m.getReplicas(),
		},
	}
	return deployment
}

func (m *stateMachine) buildPodSpec() corev1.PodSpec {
	secretVolumes, secretVolumeMounts := m.buildDeploymentSecretVolumes()

	return corev1.PodSpec{
		Volumes: append(m.getVolumes(), secretVolumes...),
		Containers: []corev1.Container{
			{
				Name:       fmt.Sprintf("%s-function-pod", m.state.instance.Name),
				Image:      m.getRuntimeImage(),
				WorkingDir: m.getWorkingSourcesDir(),
				Command: []string{
					"sh",
					"-c",
					m.getRuntimeCommand(),
				},
				Resources:    m.getResourceConfiguration(),
				Env:          m.getEnvs(),
				VolumeMounts: append(m.getVolumeMounts(), secretVolumeMounts...),
				Ports: []corev1.ContainerPort{
					{
						ContainerPort: 80,
					},
				},
			},
		},
	}
}

func (m *stateMachine) getReplicas() *int32 {
	replicas := &m.state.instance.Spec.Replicas
	if replicas != nil {
		return *replicas
	}
	defaultValue := DefaultDeploymentReplicas
	return &defaultValue
}

func (m *stateMachine) getVolumes() []corev1.Volume {
	runtime := m.state.instance.Spec.Runtime
	volumes := []corev1.Volume{
		{
			// used for writing sources (code&deps) to the sources dir
			Name: "sources",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
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

func (m *stateMachine) getVolumeMounts() []corev1.VolumeMount {
	runtime := m.state.instance.Spec.Runtime
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "sources",
			MountPath: m.getWorkingSourcesDir(),
		},
	}
	if runtime == serverlessv1alpha2.Python312 {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "local",
			MountPath: "/.local",
		})
	}
	return volumeMounts
}

func (m *stateMachine) getRuntimeImage() string {
	runtimeOverride := m.state.instance.Spec.RuntimeImageOverride
	if runtimeOverride != "" {
		return runtimeOverride
	}

	switch m.state.instance.Spec.Runtime {
	case serverlessv1alpha2.NodeJs20:
		return m.functionConfig.ImageNodeJs20
	case serverlessv1alpha2.Python312:
		return m.functionConfig.ImagePython312
	default:
		return ""
	}
}

func (m *stateMachine) getWorkingSourcesDir() string {
	switch m.state.instance.Spec.Runtime {
	case serverlessv1alpha2.NodeJs20:
		return "/usr/src/app/function"
	case serverlessv1alpha2.Python312:
		return "/kubeless"
	default:
		return ""
	}
}

func (m *stateMachine) getRuntimeCommand() string {
	spec := &m.state.instance.Spec
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

func (m *stateMachine) getEnvs() []corev1.EnvVar {
	spec := &m.state.instance.Spec
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

func (m *stateMachine) getResourceConfiguration() corev1.ResourceRequirements {
	resCfg := m.state.instance.Spec.ResourceConfiguration
	if resCfg != nil && resCfg.Function != nil && resCfg.Function.Resources != nil {
		return *resCfg.Function.Resources
	}
	return corev1.ResourceRequirements{}
}

func (m *stateMachine) buildDeploymentSecretVolumes() (volumes []corev1.Volume, volumeMounts []corev1.VolumeMount) {
	volumes = []corev1.Volume{}
	volumeMounts = []corev1.VolumeMount{}
	for _, secretMount := range m.state.instance.Spec.SecretMounts {
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
