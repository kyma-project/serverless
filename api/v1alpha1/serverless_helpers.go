package v1alpha1

import "k8s.io/utils/pointer"

const (
	defaultRegistryAddress        = "k3d-kyma-registry:5000"
	defaultServerAddress          = "k3d-kyma-registry:5000"
	defaultTraceCollectorEndpoint = "http://telemetry-otlp-traces.kyma-system.svc.cluster.local:4318/v1/traces"
	defaultPublisherProxyAddress  = "http://eventing-publisher-proxy.kyma-system.svc.cluster.local/publish"
)

func (s *ServerlessSpec) Default() {

	// if DockerRegistry struct is nil configure use of k3d registry
	if s.DockerRegistry == nil {
		s.DockerRegistry = newK3DDockerRegistry()
	}
	if s.TraceCollector == nil {
		s.TraceCollector = newTraceCollector()
	}
	if s.PublisherProxy == nil {
		s.PublisherProxy = newPublisherProxy()
	}

}

func newTraceCollector() *TraceCollector {
	return &TraceCollector{
		Value: pointer.String(defaultTraceCollectorEndpoint),
	}
}

func newPublisherProxy() *PublisherProxy {
	return &PublisherProxy{
		Value: pointer.String(defaultPublisherProxyAddress),
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
