package state

import (
	"context"
	"time"

	"github.com/kyma-project/manager-toolkit/installation/chart"
	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// delete serverless based on previously installed resources
func sFnDeleteResources(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	s.setState(v1alpha1.StateDeleting)
	s.instance.UpdateConditionUnknown(
		v1alpha1.ConditionTypeDeleted,
		v1alpha1.ConditionReasonDeletion,
		"Uninstalling",
	)

	if err := chart.CheckCRDOrphanResources(s.chartConfig); err != nil {
		// stop state machine with a warning and requeue reconciliation in 1min
		// warning state indicates that user intervention would fix it. Its not reconciliation error.
		s.setState(v1alpha1.StateWarning)
		s.instance.UpdateConditionFalse(
			v1alpha1.ConditionTypeDeleted,
			v1alpha1.ConditionReasonDeletionErr,
			err,
		)
		return stopWithEventualError(err)
	}

	return deleteResourcesWithFilter(ctx, r, s)
}

func deleteResourcesWithFilter(_ context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	done, err := chart.Uninstall(s.chartConfig, &chart.UninstallOpts{})
	if err != nil {
		return uninstallResourcesError(r, s, err)
	}

	if !done {
		return awaitingResourcesRemoval(s)
	}

	s.setState(v1alpha1.StateDeleting)
	s.instance.UpdateConditionTrue(
		v1alpha1.ConditionTypeDeleted,
		v1alpha1.ConditionReasonDeleted,
		"Serverless module deleted",
	)

	// if resources are ready to be deleted, remove finalizer
	return nextState(sFnRemoveFinalizer)
}

func uninstallResourcesError(r *reconciler, s *systemState, err error) (stateFn, *ctrl.Result, error) {
	r.log.Warnf("error while uninstalling resource %s: %s",
		client.ObjectKeyFromObject(&s.instance), err.Error())
	s.setState(v1alpha1.StateError)
	s.instance.UpdateConditionFalse(
		v1alpha1.ConditionTypeDeleted,
		v1alpha1.ConditionReasonDeletionErr,
		err,
	)
	return stopWithEventualError(err)
}

func awaitingResourcesRemoval(s *systemState) (stateFn, *ctrl.Result, error) {
	s.setState(v1alpha1.StateDeleting)
	s.instance.UpdateConditionUnknown(
		v1alpha1.ConditionTypeDeleted,
		v1alpha1.ConditionReasonDeletion,
		"Deleting module resources",
	)

	// wait one sec until ctrl-mngr remove finalizers from secrets
	return requeueAfter(time.Second)
}
