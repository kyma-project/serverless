package registry

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"math/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	dockerRegistryNodePort = 32_137

	//Available ports according to documentation https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport
	maxNodePort = 32_767
	minNodePort = 30_000
)

const (
	dockerRegistryService  = "serverless-docker-registry"
	dockerRegistryPortName = "http-registry"

	allNamespaces = ""
)

type nodePortFinder func() int32

type NodePortResolver struct {
	nodePortFinder
}

func NewNodePortResolver(finder nodePortFinder) *NodePortResolver {
	return &NodePortResolver{nodePortFinder: finder}
}

func (npr *NodePortResolver) ResolveDockerRegistryNodePortFn(ctx context.Context, k8sClient client.Client, namespace string) (int32, error) {
	svc, err := getService(ctx, k8sClient, namespace, dockerRegistryService)
	if err != nil {
		return 0, errors.Wrap(err, fmt.Sprintf("while checking if %s service is installed on cluster", dockerRegistryService))
	}

	if svc != nil && svc.Spec.Type == corev1.ServiceTypeNodePort {
		if isDefaultNodePortValue(svc) {
			return dockerRegistryNodePort, nil
		}
		currentNodePort := getNodePort(svc)
		return currentNodePort, nil
	}

	svcs, err := getAllNodePortServices(ctx, k8sClient)
	if err != nil {
		return 0, errors.Wrap(err, "while fetching all services from cluster")
	}

	if possibleConflict(svcs) {
		newPort, err := npr.drawEmptyPortNumber(svcs)
		if err != nil {
			return 0, errors.Wrap(err, "while drawing available port number")
		}
		return newPort, nil
	}
	return dockerRegistryNodePort, nil
}

func (npr *NodePortResolver) drawEmptyPortNumber(svcs *corev1.ServiceList) (int32, error) {
	nodePorts := map[int32]struct{}{}
	for _, svc := range svcs.Items {
		for _, port := range svc.Spec.Ports {
			nodePorts[port.NodePort] = struct{}{}
		}
	}

	retries := 100
	var emptyPort int32
	for i := 0; i < retries; i++ {
		possibleEmptyPort := npr.nodePortFinder()
		if _, ok := nodePorts[possibleEmptyPort]; !ok {
			emptyPort = possibleEmptyPort
			break
		}
	}
	if emptyPort == 0 {
		return 0, errors.New("couldn't draw available port number, try again")
	}
	return emptyPort, nil
}

func getNodePort(svc *corev1.Service) int32 {
	for _, port := range svc.Spec.Ports {
		if port.Name == dockerRegistryPortName {
			return port.NodePort
		}
	}
	return dockerRegistryNodePort
}

func getService(ctx context.Context, k8sClient client.Client, namespace, name string) (*corev1.Service, error) {
	svc := corev1.Service{}
	err := k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &svc)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrap(err, fmt.Sprintf("while getting %s servicce", name))
	}
	return &svc, nil
}

func isDefaultNodePortValue(svc *corev1.Service) bool {
	ports := svc.Spec.Ports
	for _, port := range ports {
		if port.NodePort == dockerRegistryNodePort {
			return true
		}
	}
	return false
}

func getAllNodePortServices(ctx context.Context, k8sClient client.Client) (*corev1.ServiceList, error) {
	svcs := corev1.ServiceList{}
	err := k8sClient.List(ctx, &svcs, &client.ListOptions{Namespace: allNamespaces})
	if err != nil {
		return nil, errors.Wrap(err, "while getting list of all services")
	}
	nodePortSvcs := &corev1.ServiceList{}
	for _, svc := range svcs.Items {
		if svc.Spec.Type == corev1.ServiceTypeNodePort {
			nodePortSvcs.Items = append(nodePortSvcs.Items, svc)
		}
		if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
			for _, port := range svc.Spec.Ports {
				if port.NodePort != 0 {
					nodePortSvcs.Items = append(nodePortSvcs.Items, svc)
					break
				}
			}
		}
	}
	return nodePortSvcs, nil
}

func possibleConflict(svcs *corev1.ServiceList) bool {
	for _, svc := range svcs.Items {
		ports := svc.Spec.Ports
		for _, port := range ports {
			if port.NodePort == dockerRegistryNodePort {
				return true
			}
		}
	}
	return false
}

var _ nodePortFinder = RandomNodePort

func RandomNodePort() int32 {
	number := rand.Int31n(maxNodePort - minNodePort)
	return minNodePort + number
}
