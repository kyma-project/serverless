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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Finalizer = "serverless-manager.kyma-project.io/deletion-hook"
)

type DockerRegistry struct {
	EnableInternal        *bool   `json:"enableInternal,omitempty"`
	InternalServerAddress *string `json:"internalServerAddress,omitempty"`
	ServerAddress         *string `json:"serverAddress,omitempty"`
	RegistryAddress       *string `json:"registryAddress,omitempty"`
	Gateway               *string `json:"gateway,omitempty"`
	GatewayCert           *string `json:"gatewayCert,omitempty"`
}

type TraceCollector struct {
	Value *string `json:"value,omitempty"`
}

type PublisherProxy struct {
	Value *string `json:"value,omitempty"`
}

// ServerlessSpec defines the desired state of Serverless
type ServerlessSpec struct {
	DockerRegistry *DockerRegistry `json:"dockerRegistry,omitempty"`
	TraceCollector *TraceCollector `json:"TraceCollector,omitempty"`
	PublisherProxy *PublisherProxy `json:"PublisherProxy,omitempty"`
}

type State string

const (
	StateReady      State = "Ready"
	StateProcessing State = "Processing"
	StateError      State = "Error"
	StateDeleting   State = "Deleting"
)

type ServerlessStatus struct {
	// State signifies current state of Serverless.
	// Value can be one of ("Ready", "Processing", "Error", "Deleting").
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Processing;Deleting;Ready;Error
	State State `json:"state"`

	TraceCollectorStatus string `json:"TraceCollectorStatus"`

	PublisherProxyStatus string `json:"PublisherProxyStatus"`

	// Conditions associated with CustomStatus.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +k8s:deepcopy-gen=true

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Serverless is the Schema for the serverlesses API
type Serverless struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServerlessSpec   `json:"spec,omitempty"`
	Status ServerlessStatus `json:"status,omitempty"`
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
