package networkpolicy

import (
	"context"

	"github.com/kyma-project/serverless/tests/serverless/internal/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	networkingclient "k8s.io/client-go/kubernetes/typed/networking/v1"
)

const (
	componentLabel = "component"
)

type NetworkPolicy struct {
	name          string
	namespace     string
	networkingCli networkingclient.NetworkPolicyInterface
	log           *logrus.Entry
	spec          networkingv1.NetworkPolicySpec
}

func NewNetworkPolicy(name, namespace string, spec networkingv1.NetworkPolicySpec, networkPolicies networkingclient.NetworkPolicyInterface, log *logrus.Entry) NetworkPolicy {
	return NetworkPolicy{
		name:          name,
		namespace:     namespace,
		networkingCli: networkPolicies,
		log:           log,
		spec:          spec,
	}
}

func (n NetworkPolicy) Create(spec networkingv1.NetworkPolicySpec) error {

	networkPolicy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n.name,
			Namespace: n.namespace,
			Labels: map[string]string{
				componentLabel: n.name,
			},
		},
		Spec: spec,
	}
	_, err := n.networkingCli.Create(context.Background(), networkPolicy, metav1.CreateOptions{})
	return errors.Wrapf(err, "while creating NetworkPolicy %s in namespace %s", n.name, n.namespace)
}

func (n NetworkPolicy) Delete() error {
	return n.networkingCli.Delete(context.Background(), n.name, metav1.DeleteOptions{})
}

func (n NetworkPolicy) LogResource() error {
	policy, err := n.Get()
	if err != nil {
		return err
	}

	out, err := utils.PrettyMarshall(policy)
	if err != nil {
		return err
	}

	n.log.Infof("%s", out)
	return nil
}

func (n NetworkPolicy) Get() (*networkingv1.NetworkPolicy, error) {
	u, err := n.networkingCli.Get(context.Background(), n.name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "while getting %s", n.name)
	}

	return u, nil
}
