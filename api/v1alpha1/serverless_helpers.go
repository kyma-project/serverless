package v1alpha1

import "k8s.io/utils/pointer"

const (
	DefaultEnableInternal  = false
	DefaultRegistryAddress = "k3d-kyma-registry:5000"
	DefaultServerAddress   = "k3d-kyma-registry:5000"
)

func (s *ServerlessSpec) Default() {
	// if DockerRegistry struct is nil configure use of k3d registry
	if s.DockerRegistry == nil {
		s.DockerRegistry = &DockerRegistry{}
	}
	if s.DockerRegistry.EnableInternal == nil {
		s.DockerRegistry.EnableInternal = pointer.Bool(DefaultEnableInternal)
	}
}

func (dr *DockerRegistry) IsInternalEnabled() bool {
	if dr != nil && dr.EnableInternal != nil {
		return *dr.EnableInternal
	}

	return false
}

func (s State) IsEmpty() bool {
	if s == "" {
		return true
	}

	return false
}
