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
	// When set to true, the internal Docker registry is enabled
	EnableInternal *bool `json:"enableInternal,omitempty"`
	// Secret used for configuration of the Docker registry
	SecretName *string `json:"secretName,omitempty"`
}

type Endpoint struct {
	Endpoint string `json:"endpoint"`
}

// ServerlessSpec defines the desired state of Serverless
type ServerlessSpec struct {
	// Used Tracing endpoint
	Tracing *Endpoint `json:"tracing,omitempty"`
	// Used Eventing endpoint
	Eventing *Endpoint `json:"eventing,omitempty"`
	// Deprecated: Defines docker registry where built images should be pushed
	// This setting should be removed from a future version where Functions won't require building images.
	DockerRegistry *DockerRegistry `json:"dockerRegistry,omitempty"`
	// Deprecated: Sets a custom CPU utilization threshold for scaling Function Pods
	// This setting should be removed from a future version where Functions won't require building images.
	TargetCPUUtilizationPercentage string `json:"targetCPUUtilizationPercentage,omitempty"`
	// Sets the requeue duration for Function. By default, the Function associated with the default configuration is requeued every 5 minutes
	FunctionRequeueDuration string `json:"functionRequeueDuration,omitempty"`
	// Deprecated: Specifies the arguments passed to the Function build executor
	// This setting should be removed from a future version where Functions won't require building images.
	FunctionBuildExecutorArgs string `json:"functionBuildExecutorArgs,omitempty"`
	// Deprecated: A number of simultaneous jobs that can run at the same time. The default value is `5`
	// This setting should be removed from a future version where Functions won't require building images.
	FunctionBuildMaxSimultaneousJobs string `json:"functionBuildMaxSimultaneousJobs,omitempty"`
	// Sets the timeout for the Function health check. The default value in seconds is `10`
	HealthzLivenessTimeout string `json:"healthzLivenessTimeout,omitempty"`
	// Deprecated: Configures the default build Job preset to be used
	// This setting should be removed from a future version where Functions won't require building images.
	DefaultBuildJobPreset string `json:"defaultBuildJobPreset,omitempty"`
	// Configures the default runtime Pod preset to be used
	DefaultRuntimePodPreset string `json:"defaultRuntimePodPreset,omitempty"`
	// Sets desired log level to be used. The default value is "info"
	LogLevel string `json:"logLevel,omitempty"`
	// Sets desired log format to be used. The default value is "json"
	LogFormat string `json:"logFormat,omitempty"`
	// Enable deployment of network policies that allow serverless' essential in-cluster communication.
	EnableNetworkPolicies bool `json:"enableNetworkPolicies,omitempty"`
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

	ConditionReasonConfiguration        = ConditionReason("Configuration")
	ConditionReasonConfigurationErr     = ConditionReason("ConfigurationErr")
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
	DefaultBuildJobPreset    string `json:"defaultBuildJobPreset,omitempty"`
	DefaultRuntimePodPreset  string `json:"defaultRuntimePodPreset,omitempty"`
	LogLevel                 string `json:"logLevel,omitempty"`
	LogFormat                string `json:"logFormat,omitempty"`

	// Used registry configuration.
	// Contains registry URL or "internal"
	DockerRegistry string `json:"dockerRegistry,omitempty"`

	// State signifies current state of Serverless.
	// Value can be one of ("Ready", "Processing", "Error", "Deleting", "Warning").
	// +kubebuilder:validation:Enum=Processing;Deleting;Ready;Error;Warning
	State State `json:"state,omitempty"`

	// Served signifies that current Serverless is managed.
	// Value can be one of ("True", "False").
	// +kubebuilder:validation:Enum=True;False
	Served Served `json:"served"`

	// NetworkPoliciesEnabled signifies if NetworkPolicies dedicated for serverless are enabled.
	// Value can be one of ("True", "False").
	// +kubebuilder:validation:Enum=True;False
	NetworkPoliciesEnabled string `json:"networkPoliciesEnabled,omitempty"`

	// Conditions associated with CustomStatus.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +k8s:deepcopy-gen=true

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:categories={kyma-modules,kyma-serverless}
//+kubebuilder:printcolumn:name="Configured",type="string",JSONPath=".status.conditions[?(@.type=='Configured')].status"
//+kubebuilder:printcolumn:name="Installed",type="string",JSONPath=".status.conditions[?(@.type=='Installed')].status"
//+kubebuilder:printcolumn:name="generation",type="integer",JSONPath=".metadata.generation"
//+kubebuilder:printcolumn:name="age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="state",type="string",JSONPath=".status.state"

// Serverless is the Schema for the serverlesses API
type Serverless struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   ServerlessSpec   `json:"spec"`
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
	metav1.ListMeta `json:"metadata"`
	Items           []Serverless `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Serverless{}, &ServerlessList{})
}
