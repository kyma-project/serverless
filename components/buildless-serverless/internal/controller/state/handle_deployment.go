package state

import (
	"context"
	"fmt"
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/internal/controller/fsm"
	"github.com/kyma-project/serverless/internal/controller/resources"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

//TODO: Add states:
// - validate - components/serverless/internal/controllers/serverless/validation.go
// - gitSources - stateFnGitCheckSources

func sFnHandleDeployment(ctx context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	m.State.Deployment = resources.NewDeployment(&m.State.Function, &m.FunctionConfig)
	builtDeployment := m.State.Deployment.Deployment

	clusterDeployment, resultGet, errGet := getOrCreateDeployment(ctx, m, builtDeployment)
	if clusterDeployment == nil {
		//TODO: think what we should return here (in context of state machine)
		return nil, resultGet, errGet
	}

	requeueNeeded, errUpdate := updateDeploymentIfNeeded(ctx, m, clusterDeployment, builtDeployment)
	if errUpdate != nil {
		//TODO: think what we should return here (in context of state machine)
		return nil, nil, errUpdate
	}
	if requeueNeeded {
		return requeue()
	}
	return nextState(sFnHandleService)
}

func getOrCreateDeployment(ctx context.Context, m *fsm.StateMachine, builtDeployment *appsv1.Deployment) (*appsv1.Deployment, *ctrl.Result, error) {
	currentDeployment := &appsv1.Deployment{}
	f := m.State.Function
	deploymentErr := m.Client.Get(ctx, client.ObjectKey{
		Namespace: f.GetNamespace(),
		Name:      f.GetName(),
	}, currentDeployment)

	if deploymentErr == nil {
		return currentDeployment, nil, nil
	}
	if !errors.IsNotFound(deploymentErr) {
		m.Log.Error(deploymentErr, "unable to fetch Deployment for Function")
		return nil, nil, deploymentErr
	}

	createResult, createErr := createDeployment(ctx, m, builtDeployment)
	return nil, createResult, createErr
}

func createDeployment(ctx context.Context, m *fsm.StateMachine, deployment *appsv1.Deployment) (*ctrl.Result, error) {
	m.Log.Info("creating a new Deployment", "Deployment.Namespace", deployment.GetNamespace(), "Deployment.Name", deployment.GetName())

	// Set the ownerRef for the Deployment, ensuring that the Deployment
	// will be deleted when the Function CR is deleted.
	controllerutil.SetControllerReference(&m.State.Function, deployment, m.Scheme)

	if err := m.Client.Create(ctx, deployment); err != nil {
		m.Log.Error(err, "failed to create new Deployment", "Deployment.Namespace", deployment.GetNamespace(), "Deployment.Name", deployment.GetName())
		m.State.Function.UpdateCondition(
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonDeploymentFailed,
			fmt.Sprintf("Deployment %s create failed: %s", deployment.GetName(), err.Error()))
		return nil, err
	}
	m.State.Function.UpdateCondition(
		serverlessv1alpha2.ConditionRunning,
		metav1.ConditionUnknown,
		serverlessv1alpha2.ConditionReasonDeploymentCreated,
		fmt.Sprintf("Deployment %s created", deployment.GetName()))

	return &ctrl.Result{RequeueAfter: time.Minute}, nil
}

func updateDeploymentIfNeeded(ctx context.Context, m *fsm.StateMachine, clusterDeployment *appsv1.Deployment, builtDeployment *appsv1.Deployment) (bool, error) {
	// Ensure the Deployment data matches the desired state
	if !deploymentChanged(clusterDeployment, builtDeployment) {
		return false, nil
	}

	//TODO: think if it's better to update only some fields
	clusterDeployment.Spec.Template = builtDeployment.Spec.Template
	clusterDeployment.Spec.Replicas = builtDeployment.Spec.Replicas
	return updateDeployment(ctx, m, clusterDeployment)
}

func deploymentChanged(a *appsv1.Deployment, b *appsv1.Deployment) bool {
	if len(a.Spec.Template.Spec.Containers) != 1 ||
		len(a.Spec.Template.Spec.Containers) != 1 {
		return true
	}
	aSpec := a.Spec.Template.Spec.Containers[0]
	bSpec := b.Spec.Template.Spec.Containers[0]

	imageChanged := aSpec.Image != bSpec.Image
	labelsChanged := !reflect.DeepEqual(a.Spec.Template.ObjectMeta.Labels, b.Spec.Template.ObjectMeta.Labels)
	replicasChanged := (a.Spec.Replicas == nil && b.Spec.Replicas != nil) ||
		(a.Spec.Replicas != nil && b.Spec.Replicas == nil) ||
		(a.Spec.Replicas != nil && b.Spec.Replicas != nil && *a.Spec.Replicas != *b.Spec.Replicas)
	workingDirChanged := !reflect.DeepEqual(aSpec.WorkingDir, bSpec.WorkingDir)
	commandChanged := !reflect.DeepEqual(aSpec.Command, bSpec.Command)
	resourcesChanged := !reflect.DeepEqual(aSpec.Resources, bSpec.Resources)
	envChanged := !reflect.DeepEqual(aSpec.Env, bSpec.Env)
	volumeMountsChanged := !reflect.DeepEqual(aSpec.VolumeMounts, bSpec.VolumeMounts)
	portsChanged := !reflect.DeepEqual(aSpec.Ports, bSpec.Ports)
	return imageChanged ||
		labelsChanged ||
		replicasChanged ||
		workingDirChanged ||
		commandChanged ||
		resourcesChanged ||
		envChanged ||
		volumeMountsChanged ||
		portsChanged
}

func updateDeployment(ctx context.Context, m *fsm.StateMachine, clusterDeployment *appsv1.Deployment) (bool, error) {
	if err := m.Client.Update(ctx, clusterDeployment); err != nil {
		m.Log.Error(err, "Failed to update Deployment", "Deployment.Namespace", clusterDeployment.GetNamespace(), "Deployment.Name", clusterDeployment.GetName())
		m.State.Function.UpdateCondition(
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonDeploymentFailed,
			fmt.Sprintf("Deployment %s update failed: %s", clusterDeployment.GetName(), err.Error()))
		return false, err
	}
	m.State.Function.UpdateCondition(
		serverlessv1alpha2.ConditionRunning,
		metav1.ConditionUnknown,
		serverlessv1alpha2.ConditionReasonDeploymentUpdated,
		fmt.Sprintf("Deployment %s updated", clusterDeployment.GetName()))
	// Requeue the request to ensure the Deployment is updated
	//TODO: rethink if it's better solution
	return true, nil
}
