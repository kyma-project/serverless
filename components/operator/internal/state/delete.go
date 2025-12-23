package state

import (
	"context"
	"time"

	"github.com/kyma-project/manager-toolkit/installation/base/resource"
	"github.com/kyma-project/manager-toolkit/installation/chart"
	"github.com/kyma-project/manager-toolkit/installation/chart/action"
	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/components/operator/internal/legacy"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

func deleteResourcesWithFilter(ctx context.Context, r *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	done, err := chart.Uninstall(s.chartConfig, &chart.UninstallOpts{
		// these resources have finalizer and the serverless-ctrl-mngr will remove them from all namespaces
		UninstallFirst: resource.OrPredicates(
			resource.AndPredicates(
				resource.HasKind("ConfigMap"),
				resource.HasLabel(legacy.ConfigLabelKey, legacy.DockerfileConfigmapLabelValue),
			),
			resource.AndPredicates(
				resource.HasKind("Secret"),
				resource.HasLabel(legacy.ConfigLabelKey, legacy.RegistrySecretLabelValue),
			),
		),
		// this resource is spread in all namespaces, but serverless-ctrl-mngr is not removing them
		PostActions: []action.PostUninstall{
			action.PostUninstallWithPredicate(
				func(u unstructured.Unstructured) (bool, error) {
					return legacy.RemoveResourceFromAllNamespaces(ctx, r.client, r.log, u)
				},
				resource.AndPredicates(
					resource.HasKind("ServiceAccount"),
					resource.HasLabel(legacy.ConfigLabelKey, legacy.ServiceAccountLabelValue),
				),
			),
		},
	})
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
