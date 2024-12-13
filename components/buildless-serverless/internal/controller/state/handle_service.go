package state

import (
	"context"
	"fmt"
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

func sFnHandleService(ctx context.Context, m *stateMachine) (stateFn, *ctrl.Result, error) {
	builtService := NewServiceBuilder(m).build()

	clusterService, resultGet, errGet := m.getOrCreateService(ctx, builtService)
	if clusterService == nil {
		//TODO: think what we should return here (in context of state machine)
		return nil, resultGet, errGet
	}

	resultUpdate, errUpdate := m.updateServiceIfNeeded(ctx, clusterService, builtService)
	if errUpdate != nil {
		//TODO: think what we should return here (in context of state machine)
		return nil, resultUpdate, errUpdate
	}
	return nextState(sFnAdjustStatus)
}

func (m *stateMachine) getOrCreateService(ctx context.Context, builtService *corev1.Service) (*corev1.Service, *ctrl.Result, error) {
	currentService := &corev1.Service{}
	f := m.state.instance
	serviceErr := m.client.Get(ctx, client.ObjectKey{
		Namespace: f.Namespace,
		Name:      f.Name,
	}, currentService)

	if serviceErr == nil {
		return currentService, nil, nil
	}
	if !errors.IsNotFound(serviceErr) {
		m.log.Error(serviceErr, "unable to fetch Service for Function")
		return nil, nil, serviceErr
	}

	createResult, createErr := m.createService(ctx, builtService)
	return nil, createResult, createErr
}

func (m *stateMachine) createService(ctx context.Context, service *corev1.Service) (*ctrl.Result, error) {
	m.log.Info("creating a new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)

	// Set the ownerRef for the Service, ensuring that the Service
	// will be deleted when the Function CR is deleted.
	controllerutil.SetControllerReference(&m.state.instance, service, m.scheme)

	if err := m.client.Create(ctx, service); err != nil {
		m.log.Error(err, "failed to create new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		m.state.instance.UpdateCondition(
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonServiceFailed,
			fmt.Sprintf("Service %s/%s create failed: %s", service.Namespace, service.Name, err.Error()))
		return nil, err
	}
	m.state.instance.UpdateCondition(
		serverlessv1alpha2.ConditionRunning,
		metav1.ConditionUnknown,
		serverlessv1alpha2.ConditionReasonServiceCreated,
		fmt.Sprintf("Service %s/%s updated", service.Namespace, service.Name))

	return &ctrl.Result{RequeueAfter: time.Minute}, nil
}

func (m *stateMachine) updateServiceIfNeeded(ctx context.Context, clusterService *corev1.Service, builtService *corev1.Service) (*ctrl.Result, error) {
	// Ensure the Deployment data matches the desired state
	if !serviceChanged(clusterService, builtService) {
		return nil, nil
	}

	// manually change fields that interest us, as clusterIP is immutable
	clusterService.Spec.Ports = builtService.Spec.Ports
	clusterService.Spec.Selector = builtService.Spec.Selector
	clusterService.Spec.Type = builtService.Spec.Type
	clusterService.ObjectMeta.Labels = builtService.GetLabels()
	return m.updateService(ctx, clusterService)
}

func serviceChanged(a *corev1.Service, b *corev1.Service) bool {
	return !mapsEqual(a.Spec.Selector, b.Spec.Selector) ||
		!mapsEqual(a.Labels, b.Labels) ||
		len(a.Spec.Ports) != 1 ||
		len(b.Spec.Ports) != 1 ||
		a.Spec.Ports[0].String() != b.Spec.Ports[0].String()
}

func (m *stateMachine) updateService(ctx context.Context, clusterService *corev1.Service) (*ctrl.Result, error) {
	if err := m.client.Update(ctx, clusterService); err != nil {
		m.log.Error(err, "Failed to update Service", "Service.Namespace", clusterService.Namespace, "Service.Name", clusterService.Name)
		m.state.instance.UpdateCondition(
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonServiceFailed,
			fmt.Sprintf("Service %s/%s update failed: %s", clusterService.Namespace, clusterService.Name, err.Error()))
		return nil, err
	}
	m.state.instance.UpdateCondition(
		serverlessv1alpha2.ConditionRunning,
		metav1.ConditionUnknown,
		serverlessv1alpha2.ConditionReasonServiceUpdated,
		fmt.Sprintf("Service %s/%s updated", clusterService.Namespace, clusterService.Name))
	// Requeue the request to ensure the Deployment is updated
	//TODO: rethink if it's better solution
	return &ctrl.Result{Requeue: true}, nil
}
