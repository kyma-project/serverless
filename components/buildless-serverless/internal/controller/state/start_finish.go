package state

import (
	"context"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

//TODO: Add states:
// - validate - components/serverless/internal/controllers/serverless/validation.go
// - gitSources - stateFnGitCheckSources
// - deploymentStatus - stateFnUpdateDeploymentStatus

// TODO: Conditions to add:
//		condition := serverlessv1alpha2.Condition{
//			Type:               serverlessv1alpha2.ConditionRunning,
//			Status:             corev1.ConditionTrue,
//			LastTransitionTime: metav1.Now(),
//			Reason:             serverlessv1alpha2.ConditionReasonDeploymentReady,
//			Message:            fmt.Sprintf("Deployment %s is ready", deploymentName),
//		}

func sFnStart(ctx context.Context, m *stateMachine) (stateFn, *controllerruntime.Result, error) {
	return sFnHandleDeployment, &controllerruntime.Result{}, nil
}

func sFnFinish(ctx context.Context, m *stateMachine) (stateFn, *controllerruntime.Result, error) {
	return nil, &controllerruntime.Result{}, nil
}
