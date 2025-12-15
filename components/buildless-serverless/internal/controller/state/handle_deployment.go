package state

import (
	"context"
	"fmt"
	"reflect"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/fsm"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/resources"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func sFnHandleDeployment(ctx context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	clusterDeployments, errGet := getDeployments(ctx, m)
	if errGet != nil {
		return stopWithError(errGet)
	}

	// If there are multiple deployments, delete them because we only want one
	if len(clusterDeployments.Items) > 1 {
		return nextState(sFnDeleteDeployments)
	}

	var clusterDeployment *appsv1.Deployment
	if len(clusterDeployments.Items) == 1 {
		clusterDeployment = &clusterDeployments.Items[0]
	}
	m.State.ClusterDeployment = clusterDeployment

	m.State.BuiltDeployment = resources.NewDeployment(&m.State.Function, &m.FunctionConfig, clusterDeployment, m.State.Commit, m.State.GitAuth, "")
	builtDeployment := m.State.BuiltDeployment.Deployment

	if m.State.ClusterDeployment == nil {
		result, errCreate := createDeployment(ctx, m, builtDeployment)
		if errCreate == nil {
			m.State.Function.CopyAnnotationsToStatus()
		}
		return nil, result, errCreate
	}

	requeueNeeded, errUpdate := updateDeploymentIfNeeded(ctx, m, clusterDeployment, builtDeployment)
	if errUpdate != nil {
		return stopWithError(errUpdate)
	}
	m.State.Function.CopyAnnotationsToStatus()
	if requeueNeeded {
		return requeue()
	}
	return nextState(sFnHandleService)
}

func getDeployments(ctx context.Context, m *fsm.StateMachine) (*appsv1.DeploymentList, error) {
	deployments := &appsv1.DeploymentList{}
	f := m.State.Function
	labels := f.InternalFunctionLabels()
	err := m.Client.List(ctx, deployments, client.InNamespace(f.GetNamespace()), client.MatchingLabels(labels))
	if err != nil {
		m.Log.Error(err, "unable to fetch Deployment for Function")
		return nil, err
	}
	return deployments, nil
}

func createDeployment(ctx context.Context, m *fsm.StateMachine, deployment *appsv1.Deployment) (*ctrl.Result, error) {
	m.Log.Info("creating a new Deployment", "Deployment.Namespace", deployment.GetNamespace(), "Deployment.Name", deployment.GetName())
	name := deployment.GetName()
	if name == "" {
		name = fmt.Sprintf("%s*", deployment.GetGenerateName())
	}

	// Set the ownerRef for the Deployment, ensuring that the Deployment
	// will be deleted when the Function CR is deleted.
	if err := controllerutil.SetControllerReference(&m.State.Function, deployment, m.Scheme); err != nil {
		m.Log.Error(err, "failed to set controller reference for new Deployment", "Deployment.Namespace", deployment.GetNamespace(), "Deployment.Name", name)
		m.State.Function.UpdateCondition(
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonDeploymentFailed,
			fmt.Sprintf("Deployment %s create failed: %s", name, err.Error()))
		return nil, err
	}

	if err := m.Client.Create(ctx, deployment); err != nil {
		m.Log.Error(err, "failed to create new Deployment", "Deployment.Namespace", deployment.GetNamespace(), "Deployment.Name", name)
		m.State.Function.UpdateCondition(
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonDeploymentFailed,
			fmt.Sprintf("Deployment %s create failed: %s", name, err.Error()))
		return nil, err
	}
	m.State.Function.UpdateCondition(
		serverlessv1alpha2.ConditionRunning,
		metav1.ConditionUnknown,
		serverlessv1alpha2.ConditionReasonDeploymentCreated,
		fmt.Sprintf("Deployment %s created", deployment.GetName()))

	return &ctrl.Result{Requeue: true}, nil
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
	annotationsChanged := !reflect.DeepEqual(a.Spec.Template.ObjectMeta.Annotations, b.Spec.Template.ObjectMeta.Annotations)
	replicasChanged := (a.Spec.Replicas == nil && b.Spec.Replicas != nil) ||
		(a.Spec.Replicas != nil && b.Spec.Replicas == nil) ||
		(a.Spec.Replicas != nil && b.Spec.Replicas != nil && *a.Spec.Replicas != *b.Spec.Replicas)
	workingDirChanged := !reflect.DeepEqual(aContainer.WorkingDir, bContainer.WorkingDir)
	commandChanged := !reflect.DeepEqual(aContainer.Command, bContainer.Command)
	resourcesChanged := !equalResources(aContainer.Resources, bContainer.Resources)
	envChanged := !reflect.DeepEqual(aContainer.Env, bContainer.Env)
	volumeMountsChanged := !reflect.DeepEqual(aContainer.VolumeMounts, bContainer.VolumeMounts)
	portsChanged := !reflect.DeepEqual(aContainer.Ports, bContainer.Ports)

	//TODO: this is a temporary solution, remove it after migrating legacy serverless
	serviceAccountChanged := a.Spec.Template.Spec.ServiceAccountName != b.Spec.Template.Spec.ServiceAccountName

	return imageChanged ||
		labelsChanged ||
		annotationsChanged ||
		replicasChanged ||
		workingDirChanged ||
		commandChanged ||
		resourcesChanged ||
		envChanged ||
		volumeMountsChanged ||
		portsChanged ||
		serviceAccountChanged ||
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
	aContainer := a.Spec.Template.Spec.InitContainers[0]
	bContainer := b.Spec.Template.Spec.InitContainers[0]

	imageChanged := aContainer.Image != bContainer.Image
	workingDirChanged := !reflect.DeepEqual(aContainer.WorkingDir, bContainer.WorkingDir)
	commandChanged := !reflect.DeepEqual(aContainer.Command, bContainer.Command)
	envChanged := !reflect.DeepEqual(aContainer.Env, bContainer.Env)
	volumeMountsChanged := !reflect.DeepEqual(aContainer.VolumeMounts, bContainer.VolumeMounts)
	return imageChanged ||
		workingDirChanged ||
		commandChanged ||
		envChanged ||
		volumeMountsChanged
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
	return true, nil
}

func equalResources(a, b corev1.ResourceRequirements) bool {
	return a.Requests.Memory().Equal(*b.Requests.Memory()) &&
		a.Requests.Cpu().Equal(*b.Requests.Cpu()) &&
		a.Limits.Memory().Equal(*b.Limits.Memory()) &&
		a.Limits.Cpu().Equal(*b.Limits.Cpu())
}
