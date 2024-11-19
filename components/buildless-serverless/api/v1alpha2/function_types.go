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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Runtime specifies the name of the Function's runtime.
type Runtime string

const (
	Python312 Runtime = "python312"
	NodeJs20  Runtime = "nodejs20"
)

// FunctionSpec defines the desired state of Function.
type FunctionSpec struct {
	// Specifies the runtime of the Function. The available values are `nodejs20`, and `python312`.
	// +kubebuilder:validation:Enum=nodejs20;python312;
	Runtime Runtime `json:"runtime"`

	// Contains the Function's source code configuration.
	/*    // +kubebuilder:validation:XValidation:message="Use GitRepository or Inline source",rule="has(self.gitRepository) && !has(self.inline) || !has(self.gitRepository) && has(self.inline)" */
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
}

type Source struct {
	// Defines the Function as git-sourced. Can't be used together with **Inline**.
	// +optional
	//	GitRepository *GitRepositorySource `json:"gitRepository,omitempty"`

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

type ResourceConfiguration struct {
	// Deprecated: Specifies resources requested by the build Job's Pod.
	// This setting should be removed from a future version where Functions won't require building images.
	// +optional
	// +kubebuilder:validation:XValidation:message="Use profile or resources",rule="has(self.profile) && !has(self.resources) || !has(self.profile) && has(self.resources)"
	// +kubebuilder:validation:XValidation:message="Invalid profile, please use one of: ['local-dev','slow','normal','fast']",rule="(!has(self.profile) || self.profile in ['local-dev','slow','normal','fast'])"
	Build *ResourceRequirements `json:"build,omitempty"`

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

// FunctionStatus defines the observed state of Function.
type FunctionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

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
