package state

import (
	"context"

	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/fsm"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
)

// TODO: Remove this state function after all Functions are migrated not to use the legacy service account.
func sFnCleanupLegacyServiceAccount(ctx context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {

	deployments, err := getDeployments(ctx, m)
	if err != nil {
		m.Log.Error(err, "Failed to list Deployments for cleaning up legacy service account name")
		return nextState(sFnValidateFunction)
	}

	for _, deployment := range deployments.Items {
		// Remove reference to legacy service account name, if present
		serviceAccountName := deployment.Spec.Template.Spec.ServiceAccountName
		if serviceAccountName == "" || serviceAccountName == "default" {
			continue
		}
		m.Log.Info("Cleaning up legacy service account from Function's Deployment")
		deployment.Spec.Template.Spec.ServiceAccountName = "default"
		deployment.Spec.Template.Spec.AutomountServiceAccountToken = ptr.To(false)
		err := m.Client.Update(ctx, &deployment)
		if err != nil {
			m.Log.Error(err, "Failed to clean up legacy service account from Deployment")
		}
		m.Log.Info("Function's Deployment updated with default service account")
	}

	return nextState(sFnValidateFunction)
}
