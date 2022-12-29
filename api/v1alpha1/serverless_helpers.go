package v1alpha1

import "k8s.io/utils/pointer"

const (
	defaultRegistryAddress = "k3d-kyma-registry:5000"
	defaultServerAddress   = "k3d-kyma-registry:5000"
)

func (s *ServerlessSpec) Default() {

	// if DockerRegistry struct is nil configure use of k3d registry
	if s.DockerRegistry == nil {
		s.DockerRegistry = newK3DDockerRegistry()
	}
}

func newK3DDockerRegistry() *DockerRegistry {
	return &DockerRegistry{
		EnableInternal:  pointer.Bool(false),
		RegistryAddress: pointer.String(defaultRegistryAddress),
		ServerAddress:   pointer.String(defaultServerAddress),
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
