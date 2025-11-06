package state

import (
	"context"
	"fmt"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/fsm"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func sFnDeleteDeployments(ctx context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	m.Log.Info("Deleting duplicated deployments")

	f := m.State.Function
	labels := f.InternalFunctionLabels()
	err := m.Client.DeleteAllOf(ctx, &appsv1.Deployment{}, &client.DeleteAllOfOptions{
		ListOptions: client.ListOptions{
			LabelSelector: apilabels.SelectorFromSet(labels),
			Namespace:     f.GetNamespace(),
		},
		DeleteOptions: client.DeleteOptions{
			PropagationPolicy: ptr.To(metav1.DeletePropagationBackground),
		},
	})

	if err != nil {
		m.Log.Error(err, "Failed to delete duplicated Deployments")
		m.State.Function.UpdateCondition(
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonDeploymentDeletionFailed,
			fmt.Sprintf("Failed to delete duplicated Deployments: %s", err.Error()))
		return stopWithError(errors.Wrap(err, "while deleting duplicated deployments"))
	}

	m.State.Function.UpdateCondition(
		serverlessv1alpha2.ConditionRunning,
		metav1.ConditionFalse,
		serverlessv1alpha2.ConditionReasonDeploymentDeleted,
		"Duplicated Deployments deleted")
	return requeue()
}
