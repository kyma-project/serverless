package resources

import (
	"fmt"
	"path"
	"strings"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/config"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/git"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/ptr"
)

const DefaultDeploymentReplicas int32 = 1
const (
	istioConfigLabelKey                      = "proxy.istio.io/config"
	istioEnableHoldUntilProxyStartLabelValue = "{ \"holdApplicationUntilProxyStarts\": true }"
	istioNativeSidecarLabelKey               = "sidecar.istio.io/nativeSidecar"
)

type deployOptions func(*Deployment)

// DeployName - set the deployment name and clear the generated name
func DeployName(name string) deployOptions {
	return func(d *Deployment) {
		d.deployName = name
		d.deployGeneratedName = ""
	}
}

// DeployTrimClusterInfoLabels - get rid of internal labels like managed-by, function-name or uuid
func DeployTrimClusterInfoLabels() deployOptions {
	return func(d *Deployment) {
		internalLabels := d.function.InternalFunctionLabels()
		for key := range internalLabels {
			delete(d.functionLabels, key)
			delete(d.selectorLabels, key)
			delete(d.podLabels, key)
		}
	}
}

// DeployAppendSelectorLabels - add additional labels to the deployment's selector
func DeployAppendSelectorLabels(labels map[string]string) deployOptions {
	return func(d *Deployment) {
		for k, v := range labels {
			d.selectorLabels[k] = v
		}
	}
}

type Deployment struct {
	*appsv1.Deployment
	functionConfig      *config.FunctionConfig
	function            *serverlessv1alpha2.Function
	clusterDeployment   *appsv1.Deployment
	commit              string
	gitAuth             *git.GitAuth
	functionLabels      map[string]string
	selectorLabels      map[string]string
	podLabels           map[string]string
	deployName          string
	deployGeneratedName string
}

func NewDeployment(f *serverlessv1alpha2.Function, c *config.FunctionConfig, clusterDeployment *appsv1.Deployment, commit string, gitAuth *git.GitAuth, opts ...deployOptions) *Deployment {
	d := &Deployment{
		functionConfig:      c,
		function:            f,
		clusterDeployment:   clusterDeployment,
		commit:              commit,
		gitAuth:             gitAuth,
		functionLabels:      f.FunctionLabels(),
		selectorLabels:      f.SelectorLabels(),
		podLabels:           f.PodLabels(),
		deployName:          "",
		deployGeneratedName: fmt.Sprintf("%s-", f.Name),
	}

	for _, o := range opts {
		o(d)
	}

	d.Deployment = d.construct()
	return d
}

func (d *Deployment) construct() *appsv1.Deployment {
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:         d.deployName,
			GenerateName: d.deployGeneratedName,
			Namespace:    d.function.Namespace,
			Labels:       d.functionLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: d.selectorLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      d.podLabels,
					Annotations: d.podAnnotations(),
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

func (d *Deployment) podRunAsUserUID() *int64 {
	return ptr.To[int64](1000) // runAsUser 1000 is the most popular and standard value for non-root user
}

func (d *Deployment) podAnnotations() map[string]string {
	result := d.defaultAnnotations()
	if d.function.Spec.Annotations != nil {
		result = labels.Merge(d.function.Spec.Annotations, result)
	}

	// merge old and new annotations to allow other components to annotate functions deployment
	// for example in case when someone use `kubectl rollout restart` on it
	// before merge we need to remove annotations that are not present in the current function to allow removing them
	result = labels.Merge(d.currentAnnotationsWithoutPreviousFunctionAnnotations(), result)
	result = labels.Merge(d.annotationsRequiredByIstio(), result)

	return result
}

func (d *Deployment) defaultAnnotations() map[string]string {
	return map[string]string{
		istioConfigLabelKey: istioEnableHoldUntilProxyStartLabelValue,
	}
}

func (d *Deployment) currentAnnotationsWithoutPreviousFunctionAnnotations() map[string]string {
	previousFunctionAnnotations := d.function.Status.FunctionAnnotations
	currentAnnotations := d.currentAnnotations()
	result := make(map[string]string)
	for key := range currentAnnotations {
		if _, ok := previousFunctionAnnotations[key]; !ok {
			result[key] = currentAnnotations[key]
		}
	}
	return result
}

// allow istio to inject native sidecar (istio-proxy as init container)
// this is required for init container of git sourced functions to fetch source from git repository
func (d *Deployment) annotationsRequiredByIstio() map[string]string {
	result := make(map[string]string)
	result[istioNativeSidecarLabelKey] = "true"
	return result
}

func (d *Deployment) currentAnnotations() map[string]string {
	if d.clusterDeployment == nil {
		return map[string]string{}
	}

	return d.clusterDeployment.Spec.Template.GetAnnotations()
}

func (d *Deployment) podSpec() corev1.PodSpec {
	secretVolumes, secretVolumeMounts := d.deploymentSecretVolumes()

	return corev1.PodSpec{
		Volumes:        append(d.volumes(), secretVolumes...),
		InitContainers: d.initContainerForGitRepository(),
		Containers: []corev1.Container{
			{
				Name:       "function",
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
					ProcMount:                ptr.To(corev1.DefaultProcMount),
					ReadOnlyRootFilesystem:   ptr.To[bool](false),
					AllowPrivilegeEscalation: ptr.To[bool](false),
					RunAsNonRoot:             ptr.To[bool](true),
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
			Name:       "init",
			Image:      d.functionConfig.Images.RepoFetcher,
			WorkingDir: d.workingSourcesDir(),
			Command: []string{
				"sh",
				"-c",
				d.initContainerCommand(),
			},
			Env: d.initContainerEnvs(),
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("50m"),
					corev1.ResourceMemory: resource.MustParse("64Mi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("200m"),
					corev1.ResourceMemory: resource.MustParse("256Mi"),
				},
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

func (d *Deployment) initContainerEnvs() []corev1.EnvVar {
	envs := []corev1.EnvVar{
		{
			Name:  "APP_REPOSITORY_URL",
			Value: d.function.Spec.Source.GitRepository.URL,
		},
		{
			Name:  "APP_REPOSITORY_REFERENCE",
			Value: d.function.Spec.Source.GitRepository.Repository.Reference,
		},
		{
			Name:  "APP_REPOSITORY_COMMIT",
			Value: d.commit,
		},
		{
			Name:  "APP_DESTINATION_PATH",
			Value: "/git-repository/repo",
		},
	}
	if d.gitAuth != nil {
		envs = append(envs, d.gitAuth.GetAuthEnvs()...)
	}

	return envs
}

func (d *Deployment) initContainerCommand() string {
	gitRepo := d.function.Spec.Source.GitRepository
	var arr []string
	arr = append(arr, "/app/gitcloner")
	arr = append(arr,
		fmt.Sprintf("mkdir /git-repository/src;cp /git-repository/repo/%s/* /git-repository/src;",
			strings.Trim(gitRepo.BaseDir, "/ ")))
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
		return d.functionConfig.Images.NodeJs20
	case serverlessv1alpha2.NodeJs22:
		return d.functionConfig.Images.NodeJs22
	case serverlessv1alpha2.Python312:
		return d.functionConfig.Images.Python312
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
	var result []string
	if d.function.HasNodejsRuntime() {
		result = append(result, `echo "{}" > package.json;`)
	}
	result = append(result, `cp /git-repository/src/* .;`)
	return strings.Join(result, "\n")
}

func (d *Deployment) runtimeCommandInlineSources() string {
	var result []string
	spec := &d.function.Spec
	dependencies := spec.Source.Inline.Dependencies

	handlerName, dependenciesName := "", ""
	if d.function.HasNodejsRuntime() {
		handlerName, dependenciesName = "handler.js", "package.json"
		result = append(result, `echo "{}" > package.json;`)
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
			Name:  "FUNC_NAME",
			Value: d.function.Name,
		},
		{
			Name:  "FUNC_RUNTIME",
			Value: string(spec.Runtime),
		},
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
	if d.function.HasNodejsRuntime() {
		envs = append(envs, []corev1.EnvVar{
			{
				Name:  "HANDLER_PATH",
				Value: "./function/handler.js",
			},
		}...)
	}
	if d.function.HasPythonRuntime() {
		envs = append(envs, []corev1.EnvVar{
			{
				Name:  "PYTHONUNBUFFERED",
				Value: "TRUE",
			},
		}...)
	}
	envs = append(envs, spec.Env...) //TODO: this order is critical, should we provide option for users to override envs?
	return envs
}

func (d *Deployment) resourceConfiguration() corev1.ResourceRequirements {
	resource, _ := d.resourceConfigurationAndProfile()
	return resource
}

func (d *Deployment) ResourceProfile() string {
	_, profile := d.resourceConfigurationAndProfile()
	return profile
}

func (d *Deployment) resourceConfigurationAndProfile() (corev1.ResourceRequirements, string) {
	cfgResources := d.functionConfig.ResourceConfig.Function.Resources
	funResource := d.function.Spec.ResourceConfiguration
	if funResource != nil && funResource.Function != nil {
		profile := funResource.Function.Profile
		if profile != "" {
			if preset, ok := cfgResources.Presets[profile]; ok {
				return preset.ToResourceRequirements(), profile
			}
		}
		if funResource.Function.Resources != nil {
			return *funResource.Function.Resources, "custom"
		}
	}
	if preset, ok := cfgResources.Presets[cfgResources.DefaultPreset]; ok {
		return preset.ToResourceRequirements(), cfgResources.DefaultPreset
	}
	return corev1.ResourceRequirements{}, "custom"
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
