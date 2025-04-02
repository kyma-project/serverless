package networkpolicy

import (
	"github.com/kyma-project/serverless/tests/serverless/internal/executor"
	"github.com/sirupsen/logrus"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	networkingclient "k8s.io/client-go/kubernetes/typed/networking/v1"
)

type newNetworkPolicyStep struct {
	name            string
	namespace       string
	networkPolicies []*NetworkPolicy
	// resCli          *resources.Resource
	log *logrus.Entry
}

// Cleanup implements executor.Step.
func (n newNetworkPolicyStep) Cleanup() error {
	for _, networkPolicy := range n.networkPolicies {
		err := networkPolicy.Delete()
		if err != nil {
			return err
		}
	}
	return nil
}

// Name implements executor.Step.
func (n newNetworkPolicyStep) Name() string {
	return n.name
}

// OnError implements executor.Step.
func (n newNetworkPolicyStep) OnError() error {
	for _, networkPolicy := range n.networkPolicies {
		err := networkPolicy.LogResource()
		if err != nil {
			return err
		}
	}
	return nil
}

// Run implements executor.Step.
func (n newNetworkPolicyStep) Run() error {
	for _, networkPolicy := range n.networkPolicies {
		err := networkPolicy.Create(networkPolicy.spec)
		if err != nil {
			return err
		}
	}
	return nil
}

var _ executor.Step = newNetworkPolicyStep{}

func CreateNetworkPoliciesStep(log *logrus.Entry, name, namespace string, networkCli networkingclient.NetworkPolicyInterface) executor.Step {

	allowEgressFromMockSpec := networkingv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{
			MatchLabels: map[string]string{
				"serverless.kyma-project.io/managed-by": "function-controller",
			},
		},
		PolicyTypes: []networkingv1.PolicyType{
			networkingv1.PolicyTypeEgress,
		},
		Egress: []networkingv1.NetworkPolicyEgressRule{
			networkingv1.NetworkPolicyEgressRule{},
		},
	}

	allowIngressToMockSpec := networkingv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app.kubernetes.io/name": "eventing-publisher-proxy",
			},
		},
		PolicyTypes: []networkingv1.PolicyType{
			networkingv1.PolicyTypeIngress,
		},
		Ingress: []networkingv1.NetworkPolicyIngressRule{
			networkingv1.NetworkPolicyIngressRule{},
		},
	}

	allowEgressFromMock := NewNetworkPolicy("allow-all-egress-from-eventing-mock", "kyma-system", allowEgressFromMockSpec, networkCli, log)
	allowIngressToMock := NewNetworkPolicy("allow-all-ingress-from-eventing-mock", "kyma-system", allowIngressToMockSpec, networkCli, log)

	return newNetworkPolicyStep{
		name:      name,
		namespace: namespace,
		networkPolicies: []*NetworkPolicy{
			&allowEgressFromMock,
			&allowIngressToMock,
		},
		log: log,
	}

}
