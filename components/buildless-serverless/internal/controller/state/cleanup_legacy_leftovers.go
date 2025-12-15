package state

import (
	"context"

	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/fsm"
	ctrl "sigs.k8s.io/controller-runtime"
)

func sFnCleanupLegacyLeftovers(ctx context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {

	deployments, err := getDeployments(ctx, m)
	if err != nil {
		m.Log.Error(err, "Failed to list Deployments for cleaning up legacy service account name")
		return nextState(sFnValidateFunction)
	}
	if len(deployments.Items) > 0 {
		for _, deployment := range deployments.Items {
			// Remove reference to legacy service account name
			serviceAccountName := deployment.Spec.Template.Spec.ServiceAccountName
			if serviceAccountName == "" {
				continue
			}
			m.Log.Info("Cleaning up legacy service account from Function's Deployment")
			deployment.Spec.Template.Spec.ServiceAccountName = ""
			deployment.Spec.Template.Spec.AutomountServiceAccountToken = nil
			m.Log.Info("Updatind Deployment to use empty service account")
			err := m.Client.Update(ctx, &deployment)
			if err != nil {
				m.Log.Error(err, "Failed to clean up legacy service account from Deployment")
			}
		}
	}

	return nextState(sFnValidateFunction)
}
