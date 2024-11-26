package state

import (
	"context"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

func sFnHandleDeployment(ctx context.Context, m *stateMachine) (stateFn, *controllerruntime.Result, error) {
	builtDeployment := m.buildDeployment()

	clusterDeployment, resultGet, errGet := m.getOrCreateDeployment(ctx, builtDeployment)
	if clusterDeployment == nil {
		return nil, &resultGet, errGet
	}

	resultUpdate, errUpdate := m.updateDeploymentIfNeeded(ctx, clusterDeployment, builtDeployment)
	if errUpdate != nil {
		return nil, &resultUpdate, errUpdate
	}
	return sFnAdjustStatus, nil, nil
}

func (m *stateMachine) getOrCreateDeployment(ctx context.Context, builtDeployment *v1.Deployment) (*v1.Deployment, controllerruntime.Result, error) {
	currentDeployment := &v1.Deployment{}
	f := m.state.instance
	deploymentErr := m.client.Get(ctx, client.ObjectKey{
		Namespace: f.Namespace,
		Name:      fmt.Sprintf("%s-function-deployment", f.Name),
	}, currentDeployment)

	if deploymentErr == nil {
		return currentDeployment, controllerruntime.Result{}, nil
	}
	if !errors.IsNotFound(deploymentErr) {
		m.log.Error(deploymentErr, "unable to fetch Deployment for Function")
		return nil, controllerruntime.Result{}, deploymentErr
	}

	createResult, createErr := m.createDeployment(ctx, builtDeployment)
	return nil, createResult, createErr
}

func (m *stateMachine) createDeployment(ctx context.Context, deployment *v1.Deployment) (controllerruntime.Result, error) {
	m.log.Info("creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)

	// Set the ownerRef for the Deployment, ensuring that the Deployment
	// will be deleted when the Function CR is deleted.
	controllerutil.SetControllerReference(&m.state.instance, deployment, m.scheme)

	if err := m.client.Create(ctx, deployment); err != nil {
		m.log.Error(err, "failed to create new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		return controllerruntime.Result{}, err
	}
	return controllerruntime.Result{RequeueAfter: time.Minute}, nil
}

func (m *stateMachine) updateDeploymentIfNeeded(ctx context.Context, clusterDeployment *v1.Deployment, builtDeployment *v1.Deployment) (controllerruntime.Result, error) {
	// Ensure the Deployment data matches the desired state
	if !cmp.Equal(clusterDeployment.Spec.Template, builtDeployment.Spec.Template) ||
		*clusterDeployment.Spec.Replicas != *builtDeployment.Spec.Replicas {
		clusterDeployment.Spec.Template = builtDeployment.Spec.Template
		clusterDeployment.Spec.Replicas = builtDeployment.Spec.Replicas
		return m.updateDeployment(ctx, clusterDeployment)
	}

	return controllerruntime.Result{}, nil
}

func (m *stateMachine) updateDeployment(ctx context.Context, clusterDeployment *v1.Deployment) (controllerruntime.Result, error) {
	if err := m.client.Update(ctx, clusterDeployment); err != nil {
		m.log.Error(err, "Failed to update Deployment", "Deployment.Namespace", clusterDeployment.Namespace, "Deployment.Name", clusterDeployment.Name)
		return controllerruntime.Result{}, err
	}
	// Requeue the request to ensure the Deployment is updated
	return controllerruntime.Result{Requeue: true}, nil
}
