/*
Copyright 2024.

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

package v1alpha2

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// Runtime specifies the name of the Function's runtime.
type Runtime string

const (
	Python312 Runtime = "python312"
	NodeJs20  Runtime = "nodejs20"
	NodeJs22  Runtime = "nodejs22"
)

// FunctionSpec defines the desired state of Function.
type FunctionSpec struct {
	// Specifies the runtime of the Function. The available values are `nodejs20`, `nodejs22`, and `python312`.
	// +kubebuilder:validation:Enum=nodejs20;nodejs22;python312;
	Runtime Runtime `json:"runtime"`

	// Specifies the runtime image used instead of the default one.
	// +optional
	RuntimeImageOverride string `json:"runtimeImageOverride,omitempty"`

	// Contains the Function's source code configuration.
	// +kubebuilder:validation:XValidation:message="Use GitRepository or Inline source",rule="has(self.gitRepository) && !has(self.inline) || !has(self.gitRepository) && has(self.inline)"
	// +kubebuilder:validation:Required
	Source Source `json:"source"`

	// Specifies an array of key-value pairs to be used as environment variables for the Function.
	// You can define values as static strings or reference values from ConfigMaps or Secrets.
	// For configuration details, see the [official Kubernetes documentation](https://kubernetes.io/docs/tasks/inject-data-application/define-environment-variable-container/).
	// +kubebuilder:validation:XValidation:message="Following envs are reserved and cannot be used: ['FUNC_RUNTIME','FUNC_HANDLER','FUNC_PORT','FUNC_HANDLER_SOURCE','FUNC_HANDLER_DEPENDENCIES','MOD_NAME','NODE_PATH','PYTHONPATH']",rule="(self.all(e, !(e.name in ['FUNC_RUNTIME','FUNC_HANDLER','FUNC_PORT','MOD_NAME','NODE_PATH','PYTHONPATH'])))"
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Specifies resources requested by the Function and the build Job.
	// +optional
	ResourceConfiguration *ResourceConfiguration `json:"resourceConfiguration,omitempty"`

	// Defines the exact number of Function's Pods to run at a time.
	// If  the Function is targeted by an external scaler,
	// then the **Replicas** field is used by the relevant HorizontalPodAutoscaler to control the number of active replicas.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default:=1
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Specifies Secrets to mount into the Function's container filesystem.
	SecretMounts []SecretMount `json:"secretMounts,omitempty"`

	// Defines labels used in Deployment's PodTemplate and applied on the Function's runtime Pod.
	// +optional
	// +kubebuilder:validation:XValidation:message="Labels has key starting with serverless.kyma-project.io/ which is not allowed",rule="!(self.exists(e, e.startsWith('serverless.kyma-project.io/')))"
	// +kubebuilder:validation:XValidation:message="Label value cannot be longer than 63",rule="self.all(e, size(e)<64)"
	Labels map[string]string `json:"labels,omitempty"`
}

type Source struct {
	// Defines the Function as git-sourced. Can't be used together with **Inline**.
	// +optional
	GitRepository *GitRepositorySource `json:"gitRepository,omitempty"`

	// Defines the Function as the inline Function. Can't be used together with **GitRepository**.
	// +optional
	Inline *InlineSource `json:"inline,omitempty"`
}

type InlineSource struct {
	// Specifies the Function's full source code.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Source string `json:"source"`

	// Specifies the Function's dependencies.
	//+optional
	Dependencies string `json:"dependencies,omitempty"`
}

type GitRepositorySource struct {
	// +kubebuilder:validation:Required

	// Specifies the URL of the Git repository with the Function's code and dependencies.
	// Depending on whether the repository is public or private and what authentication method is used to access it,
	// the URL must start with the `http(s)`, `git`, or `ssh` prefix.
	URL string `json:"url"`

	// // Specifies the authentication method. Required for SSH.
	// // +optional
	// Auth *RepositoryAuth `json:"auth,omitempty"`

	// +kubebuilder:validation:XValidation:message="BaseDir is required and cannot be empty",rule="has(self.baseDir) && (self.baseDir.trim().size() != 0)"
	// +kubebuilder:validation:XValidation:message="Reference is required and cannot be empty",rule="has(self.reference) && (self.reference.trim().size() != 0)"
	Repository `json:",inline"`
}

type Repository struct {
	// Specifies the relative path to the Git directory that contains the source code
	// from which the Function is built.
	BaseDir string `json:"baseDir,omitempty"`

	// Specifies either the branch name, tag or commit revision from which the Function Controller
	// automatically fetches the changes in the Function's code and dependencies.
	Reference string `json:"reference,omitempty"`
}

type ResourceConfiguration struct {
	// Specifies resources requested by the Function's Pod.
	// +optional
	// +kubebuilder:validation:XValidation:message="Use profile or resources",rule="has(self.profile) && !has(self.resources) || !has(self.profile) && has(self.resources)"
	// +kubebuilder:validation:XValidation:message="Invalid profile, please use one of: ['XS','S','M','L','XL']",rule="(!has(self.profile) || self.profile in ['XS','S','M','L','XL'])"
	Function *ResourceRequirements `json:"function,omitempty"`
}

type ResourceRequirements struct {
	// Defines the name of the predefined set of values of the resource.
	// Can't be used together with **Resources**.
	// +optional
	Profile string `json:"profile,omitempty"`

	// Defines the amount of resources available for the Pod.
	// Can't be used together with **Profile**.
	// For configuration details, see the [official Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/).
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
}

type SecretMount struct {
	// Specifies the name of the Secret in the Function's Namespace.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:MinLength=1
	SecretName string `json:"secretName"`

	// Specifies the path within the container where the Secret should be mounted.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	MountPath string `json:"mountPath"`
}

// FunctionStatus defines the observed state of Function.
type FunctionStatus struct {
	// Specifies the **Runtime** type of the Function.
	Runtime Runtime `json:"runtime,omitempty"`
	// Specifies the image version used to build and run the Function's Pods.
	RuntimeImage string `json:"runtimeImage,omitempty"`
	// Specifies the total number of non-terminated Pods targeted by this Function.
	Replicas int32 `json:"replicas,omitempty"`
	// Specifies the Pod selector used to match Pods in the Function's Deployment.
	PodSelector string `json:"podSelector,omitempty"`
	// Specifies the preset used for the function
	FunctionResourceProfile string `json:"functionResourceProfile,omitempty"`
	// Specifies an array of conditions describing the status of the parser.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// Specifies the commit hash used to build the Function.
	Commit string `json:"commit,omitempty"`
	// Specify the repository which was used to build the function.
	Repository `json:",inline,omitempty"`

	//TODO Should we add the GitRepository URL here?
}

type ConditionType string

const (
	ConditionRunning            ConditionType = "Running"
	ConditionConfigurationReady ConditionType = "ConfigurationReady"
)

type ConditionReason string

const (
	ConditionReasonInvalidFunctionSpec     ConditionReason = "InvalidFunctionSpec"
	ConditionReasonFunctionSpecValidated   ConditionReason = "FunctionSpecValidated"
	ConditionReasonGitSourceCheckFailed    ConditionReason = "GitSourceCheckFailed"
	ConditionReasonDeploymentCreated       ConditionReason = "DeploymentCreated"
	ConditionReasonDeploymentUpdated       ConditionReason = "DeploymentUpdated"
	ConditionReasonDeploymentFailed        ConditionReason = "DeploymentFailed"
	ConditionReasonDeploymentWaiting       ConditionReason = "DeploymentWaiting"
	ConditionReasonDeploymentReady         ConditionReason = "DeploymentReady"
	ConditionReasonServiceCreated          ConditionReason = "ServiceCreated"
	ConditionReasonServiceUpdated          ConditionReason = "ServiceUpdated"
	ConditionReasonServiceFailed           ConditionReason = "ServiceFailed"
	ConditionReasonMinReplicasNotAvailable ConditionReason = "MinReplicasNotAvailable"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Function is the Schema for the functions API.
type Function struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FunctionSpec   `json:"spec,omitempty"`
	Status FunctionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FunctionList contains a list of Function.
type FunctionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Function `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Function{}, &FunctionList{})
}

func (f *Function) UpdateCondition(c ConditionType, s metav1.ConditionStatus, r ConditionReason, msg string) {
	condition := metav1.Condition{
		Type:               string(c),
		Status:             s,
		LastTransitionTime: metav1.Now(),
		Reason:             string(r),
		Message:            msg,
	}
	meta.SetStatusCondition(&f.Status.Conditions, condition)
}

const (
	FunctionNameLabel                    = "serverless.kyma-project.io/function-name"
	FunctionManagedByLabel               = "serverless.kyma-project.io/managed-by"
	FunctionControllerValue              = "function-controller"
	FunctionUUIDLabel                    = "serverless.kyma-project.io/uuid"
	FunctionResourceLabel                = "serverless.kyma-project.io/resource"
	FunctionResourceLabelDeploymentValue = "deployment"
	PodAppNameLabel                      = "app.kubernetes.io/name"
)

func (f *Function) internalFunctionLabels() map[string]string {
	intLabels := make(map[string]string, 3)

	intLabels[FunctionNameLabel] = f.GetName()
	intLabels[FunctionManagedByLabel] = FunctionControllerValue
	intLabels[FunctionUUIDLabel] = string(f.GetUID())

	return intLabels
}

func (f *Function) FunctionLabels() map[string]string {
	internalLabels := f.internalFunctionLabels()
	functionLabels := f.GetLabels()

	return labels.Merge(functionLabels, internalLabels)
}

func (f *Function) SelectorLabels() map[string]string {
	return labels.Merge(
		map[string]string{
			FunctionResourceLabel: FunctionResourceLabelDeploymentValue,
		},
		f.internalFunctionLabels(),
	)
}

func (f *Function) PodLabels() map[string]string {
	result := f.SelectorLabels()
	if f.Spec.Labels != nil {
		result = labels.Merge(f.Spec.Labels, result)
	}
	return labels.Merge(result, map[string]string{PodAppNameLabel: f.GetName()})
}

func (f *Function) HasGitSources() bool {
	return f.Spec.Source.GitRepository != nil
}

func (f *Function) HasInlineSources() bool {
	return f.Spec.Source.Inline != nil
}

func (f *Function) HasPythonRuntime() bool {
	runtime := f.Spec.Runtime
	return runtime == Python312
}

func (f *Function) HasNodejsRuntime() bool {
	runtime := f.Spec.Runtime
	return runtime == NodeJs20 || runtime == NodeJs22
}
