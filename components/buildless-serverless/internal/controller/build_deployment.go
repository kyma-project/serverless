package controller

import (
	"fmt"
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

const DefaultDeploymentReplicas int32 = 1

func buildDeployment(function *serverlessv1alpha2.Function) *appsv1.Deployment {

	labels := map[string]string{
		"app": function.Name,
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
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: buildPodSpec(function),
			},
			Replicas: getReplicas(*function),
		},
	}
	return deployment
}

func buildPodSpec(f *serverlessv1alpha2.Function) corev1.PodSpec {
	runtime := f.Spec.Runtime

	secretVolumes, secretVolumeMounts := buildDeploymentSecretVolumes(f.Spec.SecretMounts)

	return corev1.PodSpec{
		Volumes: append(getVolumes(runtime), secretVolumes...),
		Containers: []corev1.Container{
			{
				Name:       fmt.Sprintf("%s-function-pod", f.Name),
				Image:      getRuntimeImage(runtime, f.Spec.RuntimeImageOverride),
				WorkingDir: getWorkingSourcesDir(runtime),
				Command: []string{
					"sh",
					"-c",
					getRuntimeCommand(*f),
				},
				Resources:    getResourceConfiguration(*f),
				Env:          getEnvs(*f),
				VolumeMounts: append(getVolumeMounts(runtime), secretVolumeMounts...),
				Ports: []corev1.ContainerPort{
					{
						ContainerPort: 80,
					},
				},
			},
		},
	}
}

func getReplicas(f serverlessv1alpha2.Function) *int32 {
	if f.Spec.Replicas != nil {
		return f.Spec.Replicas
	}
	defaultValue := DefaultDeploymentReplicas
	return &defaultValue
}

func getVolumes(runtime serverlessv1alpha2.Runtime) []corev1.Volume {
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

func getVolumeMounts(runtime serverlessv1alpha2.Runtime) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "sources",
			MountPath: getWorkingSourcesDir(runtime),
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

func getRuntimeImage(runtime serverlessv1alpha2.Runtime, runtimeOverride string) string {
	if runtimeOverride != "" {
		return runtimeOverride
	}

	switch runtime {
	case serverlessv1alpha2.NodeJs20:
		return "europe-docker.pkg.dev/kyma-project/prod/function-runtime-nodejs20:main"
	case serverlessv1alpha2.Python312:
		return "europe-docker.pkg.dev/kyma-project/prod/function-runtime-python312:main"
	default:
		return ""
	}
}

func getWorkingSourcesDir(runtime serverlessv1alpha2.Runtime) string {
	switch runtime {
	case serverlessv1alpha2.NodeJs20:
		return "/usr/src/app/function"
	case serverlessv1alpha2.Python312:
		return "/kubeless"
	default:
		return ""
	}
}

func getRuntimeCommand(f serverlessv1alpha2.Function) string {
	runtime := f.Spec.Runtime
	dependencies := f.Spec.Source.Inline.Dependencies
	switch runtime {
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

func getEnvs(f serverlessv1alpha2.Function) []corev1.EnvVar {
	runtime := f.Spec.Runtime
	envs := []corev1.EnvVar{
		{
			Name:  "FUNC_HANDLER_SOURCE",
			Value: f.Spec.Source.Inline.Source,
		},
		{
			Name:  "FUNC_HANDLER_DEPENDENCIES",
			Value: f.Spec.Source.Inline.Dependencies,
		},
	}
	if runtime == serverlessv1alpha2.Python312 {
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
	envs = append(envs, f.Spec.Env...)
	return envs
}

func getResourceConfiguration(f serverlessv1alpha2.Function) corev1.ResourceRequirements {
	if f.Spec.ResourceConfiguration != nil && f.Spec.ResourceConfiguration.Function != nil && f.Spec.ResourceConfiguration.Function.Resources != nil {
		return *f.Spec.ResourceConfiguration.Function.Resources
	}
	return corev1.ResourceRequirements{}
}

func buildDeploymentSecretVolumes(secretMounts []serverlessv1alpha2.SecretMount) (volumes []corev1.Volume, volumeMounts []corev1.VolumeMount) {
	volumes = []corev1.Volume{}
	volumeMounts = []corev1.VolumeMount{}
	for _, secretMount := range secretMounts {
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
