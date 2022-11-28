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
	"github.com/kyma-project/module-manager/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DockerRegistry struct {
	EnableInternal        *bool   `json:"enableInternal,omitempty"`
	InternalServerAddress *string `json:"internalServerAddress,omitempty"`
	ServerAddress         *string `json:"serverAddress,omitempty"`
	RegistryAddress       *string `json:"registryAddress,omitempty"`
}

// ServerlessSpec defines the desired state of Serverless
type ServerlessSpec struct {
	DockerRegistry *DockerRegistry `json:"dockerRegistry,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Serverless is the Schema for the serverlesses API
type Serverless struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServerlessSpec `json:"spec,omitempty"`
	Status types.Status   `json:"status,omitempty"`
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

var _ types.CustomObject = &Serverless{}

func (s *Serverless) GetStatus() types.Status {
	return s.Status
}

func (s *Serverless) SetStatus(status types.Status) {
	s.Status = status
}

func (s *Serverless) ComponentName() string {
	return "serverless"
}
