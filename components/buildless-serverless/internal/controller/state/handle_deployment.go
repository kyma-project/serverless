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

func sFnHandleDeployment(ctx context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	m.State.BuiltDeployment = resources.NewDeployment(&m.State.Function, &m.FunctionConfig, m.State.Commit)
	builtDeployment := m.State.BuiltDeployment.Deployment
	//TODO: refactor this method - split get from create

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
	if err := controllerutil.SetControllerReference(&m.State.Function, deployment, m.Scheme); err != nil {
		m.Log.Error(err, "failed to set controller reference for new Deployment", "Deployment.Namespace", deployment.GetNamespace(), "Deployment.Name", deployment.GetName())
		m.State.Function.UpdateCondition(
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonDeploymentFailed,
			fmt.Sprintf("Deployment %s create failed: %s", deployment.GetName(), err.Error()))
		return nil, err
	}

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

func updateDeploymentIfNeeded(ctx context.Context, m *fsm.StateMachine, clusterDeployment *appsv1.Deployment, builtDeployment *appsv1.Deployment) (requeueNeeded bool, err error) {
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
		len(b.Spec.Template.Spec.Containers) != 1 {
		return true
	}

	aContainer := a.Spec.Template.Spec.Containers[0]
	bContainer := b.Spec.Template.Spec.Containers[0]

	imageChanged := aContainer.Image != bContainer.Image
	labelsChanged := !reflect.DeepEqual(a.Spec.Template.ObjectMeta.Labels, b.Spec.Template.ObjectMeta.Labels)
	replicasChanged := (a.Spec.Replicas == nil && b.Spec.Replicas != nil) ||
		(a.Spec.Replicas != nil && b.Spec.Replicas == nil) ||
		(a.Spec.Replicas != nil && b.Spec.Replicas != nil && *a.Spec.Replicas != *b.Spec.Replicas)
	workingDirChanged := !reflect.DeepEqual(aContainer.WorkingDir, bContainer.WorkingDir)
	commandChanged := !reflect.DeepEqual(aContainer.Command, bContainer.Command)
	resourcesChanged := !reflect.DeepEqual(aContainer.Resources, bContainer.Resources)
	envChanged := !reflect.DeepEqual(aContainer.Env, bContainer.Env)
	volumeMountsChanged := !reflect.DeepEqual(aContainer.VolumeMounts, bContainer.VolumeMounts)
	portsChanged := !reflect.DeepEqual(aContainer.Ports, bContainer.Ports)

	return imageChanged ||
		labelsChanged ||
		replicasChanged ||
		workingDirChanged ||
		commandChanged ||
		resourcesChanged ||
		envChanged ||
		volumeMountsChanged ||
		portsChanged ||
		initContainerChanged(a, b)
}

func initContainerChanged(a *appsv1.Deployment, b *appsv1.Deployment) bool {
	// there are no init containers for inline function and one init container for git function
	// when count of init containers is not equal function type has been changed
	if len(a.Spec.Template.Spec.InitContainers) > 1 ||
		len(b.Spec.Template.Spec.InitContainers) > 1 ||
		len(a.Spec.Template.Spec.InitContainers) != len(b.Spec.Template.Spec.InitContainers) {
		return true
	}
	if len(a.Spec.Template.Spec.InitContainers) == 0 {
		return false
	}
	aInitContainer := a.Spec.Template.Spec.InitContainers[0]
	bInitContainer := b.Spec.Template.Spec.InitContainers[0]

	initCommandChanged := !reflect.DeepEqual(aInitContainer.Command, bInitContainer.Command)
	initVolumeMountsChanged := !reflect.DeepEqual(aInitContainer.VolumeMounts, bInitContainer.VolumeMounts)
	return initCommandChanged ||
		initVolumeMountsChanged
}

func updateDeployment(ctx context.Context, m *fsm.StateMachine, clusterDeployment *appsv1.Deployment) (requeueNeeded bool, err error) {
	if errUpdate := m.Client.Update(ctx, clusterDeployment); errUpdate != nil {
		m.Log.Error(errUpdate, "Failed to update Deployment", "Deployment.Namespace", clusterDeployment.GetNamespace(), "Deployment.Name", clusterDeployment.GetName())
		m.State.Function.UpdateCondition(
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonDeploymentFailed,
			fmt.Sprintf("Deployment %s update failed: %s", clusterDeployment.GetName(), errUpdate.Error()))
		return false, errUpdate
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
