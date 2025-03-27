package app

import (
	"context"

	"github.com/kyma-project/serverless/tests/serverless/internal/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	networkingclient "k8s.io/client-go/kubernetes/typed/networking/v1"
)

type NetworkPolicy struct {
	name          string
	namespace     string
	networkingCli networkingclient.NetworkPolicyInterface
	log           *logrus.Entry
}

func NewNetworkPolicy(name, namespace string, networkpolicies networkingclient.NetworkPolicyInterface, log *logrus.Entry) NetworkPolicy {
	return NetworkPolicy{
		name:          name,
		namespace:     namespace,
		networkingCli: networkpolicies,
		log:           log,
	}
}

func (n NetworkPolicy) Create() error {
	//this will ensure a network policy allowing incomming trafic towards gitserver pod
	networkpolicy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: n.name,
			Labels: map[string]string{
				componentLabel: n.name,
			},
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"component": "gitserver",
				},
			},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
				networkingv1.PolicyTypeEgress,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				networkingv1.NetworkPolicyIngressRule{},
			},
			Egress: []networkingv1.NetworkPolicyEgressRule{
				networkingv1.NetworkPolicyEgressRule{},
			},
		},
	}
	_, err := n.networkingCli.Create(context.Background(), networkpolicy, metav1.CreateOptions{})
	return errors.Wrapf(err, "while creating NetworkPolicy %s in namespace %s", n.name, n.namespace)
}

func (n NetworkPolicy) Delete(ctx context.Context, options metav1.DeleteOptions) error {
	return n.networkingCli.Delete(ctx, n.name, options)
}

func (n NetworkPolicy) Get(ctx context.Context, options metav1.GetOptions) (*networkingv1.NetworkPolicy, error) {
	networkpolicy, err := n.networkingCli.Get(ctx, n.name, options)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting network policy %s in namespace %s", n.name, n.namespace)
	}
	return networkpolicy, nil
}
func (n NetworkPolicy) LogResource() error {
	networkpolicy, err := n.Get(context.TODO(), metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "while getting network policy")
	}
	out, err := utils.PrettyMarshall(networkpolicy)
	if err != nil {
		return errors.Wrap(err, "while marshalling network policy")
	}
	n.log.Info(out)
	return nil
}
