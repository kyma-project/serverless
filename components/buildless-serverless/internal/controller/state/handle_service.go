package state

import (
	"context"
	"fmt"
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/internal/controller/fsm"
	"github.com/kyma-project/serverless/internal/controller/resources"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

func sFnHandleService(ctx context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	builtService := resources.NewService(&m.State.Function).Service

	clusterService, resultGet, errGet := getOrCreateService(ctx, m, builtService)
	if clusterService == nil {
		//TODO: think what we should return here (in context of state machine)
		return nil, resultGet, errGet
	}

	resultUpdate, errUpdate := updateServiceIfNeeded(ctx, m, clusterService, builtService)
	if errUpdate != nil {
		//TODO: think what we should return here (in context of state machine)
		return nil, resultUpdate, errUpdate
	}
	return nextState(sFnDeploymentStatus)
}

func getOrCreateService(ctx context.Context, m *fsm.StateMachine, builtService *corev1.Service) (*corev1.Service, *ctrl.Result, error) {
	currentService := &corev1.Service{}
	f := m.State.Function
	serviceErr := m.Client.Get(ctx, client.ObjectKey{
		Namespace: f.GetNamespace(),
		Name:      f.GetName(),
	}, currentService)

	if serviceErr == nil {
		return currentService, nil, nil
	}
	if !errors.IsNotFound(serviceErr) {
		m.Log.Error(serviceErr, "unable to fetch Service for Function")
		return nil, nil, serviceErr
	}

	createResult, createErr := createService(ctx, m, builtService)
	return nil, createResult, createErr
}

func createService(ctx context.Context, m *fsm.StateMachine, service *corev1.Service) (*ctrl.Result, error) {
	m.Log.Info("creating a new Service", "Service.Namespace", service.GetNamespace(), "Service.Name", service.GetName())

	// Set the ownerRef for the Service, ensuring that the Service
	// will be deleted when the Function CR is deleted.
	if err := controllerutil.SetControllerReference(&m.State.Function, service, m.Scheme); err != nil {
		m.Log.Error(err, "failed to set controller reference for new Service", "Service.Namespace", service.GetNamespace(), "Service.Name", service.GetName())
		m.State.Function.UpdateCondition(
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonServiceFailed,
			fmt.Sprintf("Service %s create failed: %s", service.GetName(), err.Error()))
		return nil, err
	}

	if err := m.Client.Create(ctx, service); err != nil {
		m.Log.Error(err, "failed to create new Service", "Service.Namespace", service.GetNamespace(), "Service.Name", service.GetName())
		m.State.Function.UpdateCondition(
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonServiceFailed,
			fmt.Sprintf("Service %s create failed: %s", service.GetName(), err.Error()))
		return nil, err
	}
	m.State.Function.UpdateCondition(
		serverlessv1alpha2.ConditionRunning,
		metav1.ConditionUnknown,
		serverlessv1alpha2.ConditionReasonServiceCreated,
		fmt.Sprintf("Service %s created", service.GetName()))

	return &ctrl.Result{RequeueAfter: time.Minute}, nil
}

func updateServiceIfNeeded(ctx context.Context, m *fsm.StateMachine, clusterService *corev1.Service, builtService *corev1.Service) (*ctrl.Result, error) {
	// Ensure the Deployment data matches the desired state
	if !serviceChanged(clusterService, builtService) {
		return nil, nil
	}

	// manually change fields that interest us, as clusterIP is immutable
	clusterService.Spec.Ports = builtService.Spec.Ports
	clusterService.Spec.Selector = builtService.Spec.Selector
	clusterService.Spec.Type = builtService.Spec.Type
	clusterService.ObjectMeta.Labels = builtService.GetLabels()
	return updateService(ctx, m, clusterService)
}

func serviceChanged(a *corev1.Service, b *corev1.Service) bool {
	return !mapsEqual(a.Spec.Selector, b.Spec.Selector) ||
		!mapsEqual(a.Labels, b.Labels) ||
		len(a.Spec.Ports) != 1 ||
		len(b.Spec.Ports) != 1 ||
		a.Spec.Ports[0].String() != b.Spec.Ports[0].String()
}

func updateService(ctx context.Context, m *fsm.StateMachine, clusterService *corev1.Service) (*ctrl.Result, error) {
	if err := m.Client.Update(ctx, clusterService); err != nil {
		m.Log.Error(err, "Failed to update Service", "Service.Namespace", clusterService.GetNamespace(), "Service.Name", clusterService.GetName())
		m.State.Function.UpdateCondition(
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonServiceFailed,
			fmt.Sprintf("Service %s update failed: %s", clusterService.GetName(), err.Error()))
		return nil, err
	}
	m.State.Function.UpdateCondition(
		serverlessv1alpha2.ConditionRunning,
		metav1.ConditionUnknown,
		serverlessv1alpha2.ConditionReasonServiceUpdated,
		fmt.Sprintf("Service %s updated", clusterService.GetName()))
	// Requeue the request to ensure the Deployment is updated
	//TODO: rethink if it's better solution
	return &ctrl.Result{Requeue: true}, nil
}
