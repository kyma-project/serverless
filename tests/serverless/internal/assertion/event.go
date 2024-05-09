package assertion

import (
	"context"
	"github.com/kyma-project/serverless/tests/serverless/internal/executor"
	"github.com/kyma-project/serverless/tests/serverless/internal/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type eventsWarningCheck struct {
	name      string
	namespace string
	coreCli   typedcorev1.CoreV1Interface
	log       *logrus.Entry
}

func EventWarningCheck(l *logrus.Entry, name, namespace string, coreCli typedcorev1.CoreV1Interface) executor.Step {
	return eventsWarningCheck{
		name:      name,
		namespace: namespace,
		log:       l,
		coreCli:   coreCli,
	}
}

func (e eventsWarningCheck) Name() string {
	return e.name
}

func (e eventsWarningCheck) Run() error {
	typeSelector := fields.OneTermEqualSelector("type", corev1.EventTypeWarning)
	nsSelector := fields.OneTermEqualSelector("metadata.namespace", e.namespace)
	selector := fields.AndSelectors(typeSelector, nsSelector).String()

	list, err := e.coreCli.Events(e.namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: selector,
	})
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}
	if len(list.Items) != 0 {
		out, err := utils.PrettyMarshall(list)
		if err != nil {
			return err
		}
		return errors.Errorf("Found warning events in namespace:%s, events: %s", e.namespace, out)
	}
	return nil
}

func (e eventsWarningCheck) Cleanup() error {
	return nil
}

func (e eventsWarningCheck) OnError() error {
	return nil
}

var _ executor.Step = eventsWarningCheck{}
