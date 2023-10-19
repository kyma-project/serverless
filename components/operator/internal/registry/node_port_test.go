package registry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const nonConflictPort int32 = 32238

const kymaNamespace = "kyma-system"

type assertFn func(t *testing.T, overrides map[string]interface{})

func TestNodePortAction(t *testing.T) {
	testCases := map[string]struct {
		givenService *corev1.Service
		expectedPort int32
		assertFn     assertFn
	}{
		"Return default port new port when nodePort installed on default port": {
			givenService: fixtureServiceNodePort(dockerRegistryService, kymaNamespace, dockerRegistryNodePort),
			expectedPort: dockerRegistryNodePort,
		},
		"Generate new port when nodePort service installed on different port": {
			givenService: fixtureServiceNodePort(dockerRegistryService, kymaNamespace, nonConflictPort),
			expectedPort: nonConflictPort,
		},
		"Return default port new port when nodePort not installed, without port conflict": {
			expectedPort: dockerRegistryNodePort,
		},
		"Generate new port when nodePort not installed, with port conflict": {
			givenService: fixtureServiceNodePort("conflicting-svc", kymaNamespace, dockerRegistryNodePort),
			expectedPort: nonConflictPort,
		},
		"Return default port new port when service is ClusterIP before upgrade without port conflict": {
			givenService: fixtureServiceClusterIP(dockerRegistryService, kymaNamespace),
			expectedPort: dockerRegistryNodePort,
		},
		"Generate new port when cluster has NodePort service in different namespace with port conflict": {
			givenService: fixtureServiceNodePort(dockerRegistryService, "different-ns", dockerRegistryNodePort),
			expectedPort: nonConflictPort,
		},
		"Generate new port when cluster has LoadBalancer service in different namespace with port conflict": {
			givenService: fixtureLoadBalancer(),
			expectedPort: nonConflictPort,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			//GIVEN
			ctx := context.TODO()
			k8sClient := fake.NewClientBuilder().
				WithRuntimeObjects(fixtureServices()...).
				Build()
			resolver := NewNodePortResolver(fixedNodePort(nonConflictPort))
			if testCase.givenService != nil {
				err := k8sClient.Create(ctx, testCase.givenService, &client.CreateOptions{})
				require.NoError(t, err)
			}

			//WHEN
			port, err := resolver.ResolveDockerRegistryNodePortFn(ctx, k8sClient, kymaNamespace)

			//THEN
			require.NoError(t, err)
			require.Equal(t, testCase.expectedPort, port)
		})
	}
}

func fixtureServiceNodePort(name, namespace string, nodePort int32) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				{Name: dockerRegistryPortName, NodePort: nodePort}},
		},
	}
}

func fixtureServiceClusterIP(name, namespace string) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{Name: dockerRegistryPortName, Port: 5000}},
		},
	}
}

func fixtureServices() []runtime.Object {
	l := []runtime.Object{
		fixtureServiceNodePort("other-node-port", kymaNamespace, dockerRegistryNodePort-1),
		fixtureServiceNodePort("many-ports", kymaNamespace, dockerRegistryNodePort+2),
	}
	return l
}

func fixedNodePort(expectedPort int32) func() int32 {
	return func() int32 {
		return expectedPort
	}
}

func fixtureLoadBalancer() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio-ingressgateway",
			Namespace: "istio-system",
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeLoadBalancer,
			Ports: []corev1.ServicePort{
				{
					NodePort: dockerRegistryNodePort,
					Name:     "http2",
				},
				{
					NodePort: 30857,
					Name:     "https",
				},
			},
		},
		Status: corev1.ServiceStatus{},
	}
}
