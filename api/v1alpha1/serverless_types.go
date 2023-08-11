/*
Copyright 2022.

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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DockerRegistry struct {
	EnableInternal *bool   `json:"enableInternal,omitempty"`
	SecretName     *string `json:"secretName,omitempty"`
}

type Endpoint struct {
	Endpoint string `json:"endpoint"`
}

// ServerlessSpec defines the desired state of Serverless
type ServerlessSpec struct {
	Tracing                          *Endpoint       `json:"tracing,omitempty"`
	Eventing                         *Endpoint       `json:"eventing,omitempty"`
	DockerRegistry                   *DockerRegistry `json:"dockerRegistry,omitempty"`
	TargetCPUUtilizationPercentage   string          `json:"targetCPUUtilizationPercentage,omitempty"`
	FunctionRequeueDuration          string          `json:"functionRequeueDuration,omitempty"`
	FunctionBuildExecutorArgs        string          `json:"functionBuildExecutorArgs,omitempty"`
	FunctionBuildMaxSimultaneousJobs string          `json:"functionBuildMaxSimultaneousJobs,omitempty"`
	HealthzLivenessTimeout           string          `json:"healthzLivenessTimeout,omitempty"`
	FunctionRequestBodyLimitMb       string          `json:"functionRequestBodyLimitMb,omitempty"`
	FunctionTimeoutSec               string          `json:"functionTimeoutSec,omitempty"`
	DefaultBuildJobPreset            string          `json:"defaultBuildJobPreset,omitempty"`
	DefaultRuntimePodPreset          string          `json:"defaultRuntimePodPreset,omitempty"`
}

type State string

type Served string

type ConditionReason string

type ConditionType string

const (
	StateReady      State = "Ready"
	StateProcessing State = "Processing"
	StateWarning    State = "Warning"
	StateError      State = "Error"
	StateDeleting   State = "Deleting"

	ServedTrue  Served = "True"
	ServedFalse Served = "False"

	// installation and deletion details
	ConditionTypeInstalled = ConditionType("Installed")

	// prerequisites and soft dependencies
	ConditionTypeConfigured = ConditionType("Configured")

	// deletion
	ConditionTypeDeleted = ConditionType("Deleted")

	ConditionReasonConfigurationCheck   = ConditionReason("ConfigurationCheck")
	ConditionReasonConfigurationErr     = ConditionReason("ConfigurationCheckErr")
	ConditionReasonConfigured           = ConditionReason("Configured")
	ConditionReasonInstallation         = ConditionReason("Installation")
	ConditionReasonInstallationErr      = ConditionReason("InstallationErr")
	ConditionReasonInstalled            = ConditionReason("Installed")
	ConditionReasonServerlessDuplicated = ConditionReason("ServerlessDuplicated")
	ConditionReasonDeletion             = ConditionReason("Deletion")
	ConditionReasonDeletionErr          = ConditionReason("DeletionErr")
	ConditionReasonDeleted              = ConditionReason("Deleted")

	Finalizer = "serverless-operator.kyma-project.io/deletion-hook"
)

type ServerlessStatus struct {
	// Used the Eventing endpoint and the Tracing endpoint.
	EventingEndpoint string `json:"eventingEndpoint,omitempty"`
	TracingEndpoint  string `json:"tracingEndpoint,omitempty"`

	CPUUtilizationPercentage string `json:"targetCPUUtilizationPercentage,omitempty"`
	RequeueDuration          string `json:"functionRequeueDuration,omitempty"`
	BuildExecutorArgs        string `json:"functionBuildExecutorArgs,omitempty"`
	BuildMaxSimultaneousJobs string `json:"functionBuildMaxSimultaneousJobs,omitempty"`
	HealthzLivenessTimeout   string `json:"healthzLivenessTimeout,omitempty"`
	RequestBodyLimitMb       string `json:"functionRequestBodyLimitMb,omitempty"`
	TimeoutSec               string `json:"functionTimeoutSec,omitempty"`
	DefaultBuildJobPreset    string `json:"defaultBuildJobPreset,omitempty"`
	DefaultRuntimePodPreset  string `json:"defaultRuntimePodPreset,omitempty"`

	// Used registry configuration.
	// Contains registry URL or "internal"
	DockerRegistry string `json:"dockerRegistry,omitempty"`

	// State signifies current state of Serverless.
	// Value can be one of ("Ready", "Processing", "Error", "Deleting").
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Processing;Deleting;Ready;Error;Warning
	State State `json:"state,omitempty"`

	// Served signifies that current Serverless is managed.
	// Value can be one of ("True", "False").
	// +kubebuilder:validation:Enum=True;False
	Served Served `json:"served"`

	// Conditions associated with CustomStatus.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +k8s:deepcopy-gen=true

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Configured",type="string",JSONPath=".status.conditions[?(@.type=='Configured')].status"
//+kubebuilder:printcolumn:name="Installed",type="string",JSONPath=".status.conditions[?(@.type=='Installed')].status"
//+kubebuilder:printcolumn:name="generation",type="integer",JSONPath=".metadata.generation"
//+kubebuilder:printcolumn:name="age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="state",type="string",JSONPath=".status.state"

// Serverless is the Schema for the serverlesses API
type Serverless struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServerlessSpec   `json:"spec,omitempty"`
	Status ServerlessStatus `json:"status,omitempty"`
}

func (s *Serverless) UpdateConditionFalse(c ConditionType, r ConditionReason, err error) {
	condition := metav1.Condition{
		Type:               string(c),
		Status:             "False",
		LastTransitionTime: metav1.Now(),
		Reason:             string(r),
		Message:            err.Error(),
	}
	meta.SetStatusCondition(&s.Status.Conditions, condition)
}

func (s *Serverless) UpdateConditionUnknown(c ConditionType, r ConditionReason, msg string) {
	condition := metav1.Condition{
		Type:               string(c),
		Status:             "Unknown",
		LastTransitionTime: metav1.Now(),
		Reason:             string(r),
		Message:            msg,
	}
	meta.SetStatusCondition(&s.Status.Conditions, condition)
}

func (s *Serverless) UpdateConditionTrue(c ConditionType, r ConditionReason, msg string) {
	condition := metav1.Condition{
		Type:               string(c),
		Status:             "True",
		LastTransitionTime: metav1.Now(),
		Reason:             string(r),
		Message:            msg,
	}
	meta.SetStatusCondition(&s.Status.Conditions, condition)
}

func (s *Serverless) IsServedEmpty() bool {
	return s.Status.Served == ""
}

//+kubebuilder:object:root=true

// ServerlessList contains a list of Serverless
type ServerlessList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Serverless `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Serverless{}, &ServerlessList{})
}
