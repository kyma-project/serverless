package resources

import (
	"fmt"
	"path"
	"strings"

	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/internal/config"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

const DefaultDeploymentReplicas int32 = 1

type Deployment struct {
	*appsv1.Deployment
	functionConfig *config.FunctionConfig
	function       *serverlessv1alpha2.Function
	commit         string
}

func NewDeployment(f *serverlessv1alpha2.Function, c *config.FunctionConfig, commit string) *Deployment {
	d := &Deployment{
		functionConfig: c,
		function:       f,
		commit:         commit,
	}
	d.Deployment = d.construct()
	return d
}

func (d *Deployment) construct() *appsv1.Deployment {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.name(),
			Namespace: d.function.Namespace,
			Labels:    d.function.FunctionLabels(),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: d.function.SelectorLabels(),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: d.function.PodLabels(),
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

func (d *Deployment) podRunAsUserUID() *int64 {
	return ptr.To[int64](1000) // runAsUser 1000 is the most popular and standard value for non-root user
}

func (d *Deployment) podSpec() corev1.PodSpec {
	secretVolumes, secretVolumeMounts := d.deploymentSecretVolumes()

	return corev1.PodSpec{
		Volumes:        append(d.volumes(), secretVolumes...),
		InitContainers: d.initContainerForGitRepository(),
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
						ContainerPort: 8080,
						Protocol:      "TCP",
					},
				},
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
				SecurityContext: &corev1.SecurityContext{
					Privileged: ptr.To[bool](false),
					Capabilities: &corev1.Capabilities{
						Drop: []corev1.Capability{
							"ALL",
						},
					},
					ProcMount:              ptr.To(corev1.DefaultProcMount),
					ReadOnlyRootFilesystem: ptr.To[bool](false),
				},
			},
		},
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser:  d.podRunAsUserUID(),
			RunAsGroup: d.podRunAsUserUID(),
			SeccompProfile: &corev1.SeccompProfile{
				Type: corev1.SeccompProfileTypeRuntimeDefault,
			},
		},
	}
}

func (d *Deployment) initContainerForGitRepository() []corev1.Container {
	if !d.function.HasGitSources() {
		return []corev1.Container{}
	}
	return []corev1.Container{
		{
			Name: fmt.Sprintf("%s-init", d.name()),
			//TODO: should we use this image?
			Image:      "europe-docker.pkg.dev/kyma-project/prod/alpine-git:v20250212-39c86988",
			WorkingDir: d.workingSourcesDir(),
			Command: []string{
				"sh",
				"-c",
				d.initContainerCommand(),
			},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "git-repository",
					ReadOnly:  false,
					MountPath: "/git-repository",
				},
			},
			SecurityContext: &corev1.SecurityContext{
				Privileged: ptr.To[bool](false),
				Capabilities: &corev1.Capabilities{
					Drop: []corev1.Capability{
						"ALL",
					},
				},
				ProcMount:              ptr.To(corev1.DefaultProcMount),
				ReadOnlyRootFilesystem: ptr.To[bool](false),
			},
		},
	}
}

func (d *Deployment) initContainerCommand() string {
	gitRepo := d.function.Spec.Source.GitRepository
	var arr []string
	arr = append(arr,
		fmt.Sprintf("git clone --depth 1 --branch %s %s /git-repository/repo;", gitRepo.Reference, gitRepo.URL))

	if d.commit != "" {
		arr = append(arr,
			fmt.Sprintf("cd /git-repository/repo;git reset --hard %s; cd ../..;", d.commit))
	}

	arr = append(arr,
		fmt.Sprintf("mkdir /git-repository/src;cp /git-repository/repo/%s/* /git-repository/src;", strings.Trim(gitRepo.BaseDir, "/ ")))

	return strings.Join(arr, "\n")
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
		{
			Name: "tmp",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}
	if d.function.HasGitSources() {
		volumes = append(volumes, corev1.Volume{
			Name: "git-repository",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}
	if d.function.HasPythonRuntime() {
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
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "sources",
			MountPath: d.workingSourcesDir(),
		},
		{
			Name:      "tmp",
			ReadOnly:  false,
			MountPath: "/tmp",
		},
	}
	if d.function.HasGitSources() {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "git-repository",
			MountPath: "/git-repository",
		})
	}
	if d.function.HasNodejsRuntime() {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "package-registry-config",
			MountPath: path.Join(d.workingSourcesDir(), "package-registry-config/.npmrc"),
			SubPath:   ".npmrc",
		})
	}
	if d.function.HasPythonRuntime() {
		volumeMounts = append(volumeMounts,
			corev1.VolumeMount{
				Name:      "local",
				MountPath: "/.local",
			},
			corev1.VolumeMount{
				Name:      "package-registry-config",
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
	case serverlessv1alpha2.NodeJs22:
		return d.functionConfig.ImageNodeJs22
	case serverlessv1alpha2.Python312:
		return d.functionConfig.ImagePython312
	default:
		return ""
	}
}

func (d *Deployment) workingSourcesDir() string {
	if d.function.HasNodejsRuntime() {
		return "/usr/src/app/function"
	} else if d.function.HasPythonRuntime() {
		return "/kubeless"
	}
	return ""
}

func (d *Deployment) runtimeCommand() string {
	var result []string
	result = append(result, d.runtimeCommandSources())
	result = append(result, d.runtimeCommandInstall())
	result = append(result, d.runtimeCommandStart())

	return strings.Join(result, "\n")
}

func (d *Deployment) runtimeCommandSources() string {
	spec := &d.function.Spec
	if spec.Source.GitRepository != nil {
		return d.runtimeCommandGitSources()
	}
	return d.runtimeCommandInlineSources()
}

func (d *Deployment) runtimeCommandGitSources() string {
	return "cp /git-repository/src/* .;"
}

func (d *Deployment) runtimeCommandInlineSources() string {
	var result []string
	spec := &d.function.Spec
	dependencies := spec.Source.Inline.Dependencies

	handlerName, dependenciesName := "", ""
	if d.function.HasNodejsRuntime() {
		handlerName, dependenciesName = "handler.js", "package.json"
	} else if d.function.HasPythonRuntime() {
		handlerName, dependenciesName = "handler.py", "requirements.txt"
	}

	result = append(result, fmt.Sprintf(`echo "${FUNC_HANDLER_SOURCE}" > %s;`, handlerName))
	if dependencies != "" {
		result = append(result, fmt.Sprintf(`echo "${FUNC_HANDLER_DEPENDENCIES}" > %s;`, dependenciesName))
	}
	return strings.Join(result, "\n")
}

func (d *Deployment) runtimeCommandInstall() string {
	if d.function.HasNodejsRuntime() {
		return `npm install --prefer-offline --no-audit --progress=false;`
	} else if d.function.HasPythonRuntime() {
		return `PIP_CONFIG_FILE=package-registry-config/pip.conf pip install --user --no-cache-dir -r /kubeless/requirements.txt;`
	}
	return ""
}

func (d *Deployment) runtimeCommandStart() string {
	if d.function.HasNodejsRuntime() {
		return `cd ..;
npm start;`
	} else if d.function.HasPythonRuntime() {
		return `cd ..;
python /kubeless.py;`
	}
	return ""
}

func (d *Deployment) envs() []corev1.EnvVar {
	spec := &d.function.Spec
	envs := []corev1.EnvVar{
		{
			Name:  "SERVICE_NAMESPACE",
			Value: d.function.Namespace,
		},
		{
			Name:  "TRACE_COLLECTOR_ENDPOINT",
			Value: d.functionConfig.FunctionTraceCollectorEndpoint,
		},
		{
			Name:  "PUBLISHER_PROXY_ADDRESS",
			Value: d.functionConfig.FunctionPublisherProxyAddress,
		},
	}
	if d.function.HasInlineSources() {
		envs = append(envs, []corev1.EnvVar{
			{
				Name:  "FUNC_HANDLER_SOURCE",
				Value: spec.Source.Inline.Source,
			},
			{
				Name:  "FUNC_HANDLER_DEPENDENCIES",
				Value: spec.Source.Inline.Dependencies,
			},
		}...)
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
	envs = append(envs, spec.Env...) //TODO: this order is critical, should we provide option for users to override envs?
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
