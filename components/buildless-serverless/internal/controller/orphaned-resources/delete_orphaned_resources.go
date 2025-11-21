package orphaned_resources

import (
	"context"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	apilabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"strings"
)

func DeleteOrphanedResources(ctx context.Context, m manager.Manager) error {
	m.GetLogger().Info("cleaning orphan deprecated resources")
	labels := map[string]string{
		"serverless.kyma-project.io/managed-by": "function-controller",
	}

	runtimeLabels := map[string]string{
		"serverless.kyma-project.io/config": "runtime",
		"app.kubernetes.io/part-of":         "serverless",
	}

	dockerRegistryLabels := map[string]string{
		"kyma-project.io/module": "serverless",
		"app.kubernetes.io/name": "docker-registry",
	}

	credentialsLabels := map[string]string{
		"serverless.kyma-project.io/config": "credentials",
		"app.kubernetes.io/part-of":         "serverless",
	}

	serviceAccountsName := "serverless-function"

	var collectedErrors []string

	// list orphaned jobs
	jobs := &batchv1.JobList{}
	err := listOrphanedResources(ctx, m.GetAPIReader(), jobs, labels)
	if err != nil {
		collectedErrors = append(collectedErrors, fmt.Sprintf("failed to list orphaned jobs: %s", err))
	}

	// delete orphaned jobs
	for _, job := range jobs.Items {
		err := deleteOrphanedResource(ctx, m.GetClient(), &job)
		if err != nil && !errors.IsNotFound(err) {
			collectedErrors = append(collectedErrors, fmt.Sprintf("failed to delete orphaned-resources %s/%s: %s", job.Namespace, job.Name, err))
		}
	}

	// list orphaned configmaps
	configMaps := &corev1.ConfigMapList{}
	err = listOrphanedResources(ctx, m.GetAPIReader(), configMaps, labels)
	if err != nil {
		collectedErrors = append(collectedErrors, fmt.Sprintf("failed to list orphaned configmaps: %s", err))
	}

	// delete orphaned configmaps
	for _, configMap := range configMaps.Items {
		err := deleteOrphanedResource(ctx, m.GetClient(), &configMap)
		if err != nil && !errors.IsNotFound(err) {
			collectedErrors = append(collectedErrors, fmt.Sprintf("failed to delete orphaned configmap %s/%s: %s", configMap.Namespace, configMap.Name, err))
		}
	}

	// list orphaned runtime configmaps
	runtimeConfigMaps := &corev1.ConfigMapList{}
	err = listOrphanedResources(ctx, m.GetAPIReader(), runtimeConfigMaps, runtimeLabels)
	if err != nil {
		collectedErrors = append(collectedErrors, fmt.Sprintf("failed to list orphaned runtime configmaps: %s", err))
	}

	// delete orphaned runtime configmaps
	for _, runtimeConfigMap := range runtimeConfigMaps.Items {
		err := deleteOrphanedResource(ctx, m.GetClient(), &runtimeConfigMap)
		if err != nil && !errors.IsNotFound(err) {
			collectedErrors = append(collectedErrors, fmt.Sprintf("failed to delete orphaned runtime configmap %s/%s: %s", runtimeConfigMap.Namespace, runtimeConfigMap.Name, err))
		}
	}

	// list orphaned docker registry configmaps
	dockerRegistryConfigMaps := &corev1.ConfigMapList{}
	err = listOrphanedResources(ctx, m.GetAPIReader(), dockerRegistryConfigMaps, dockerRegistryLabels)
	if err != nil {
		collectedErrors = append(collectedErrors, fmt.Sprintf("failed to list orphaned docker registry configmaps: %s", err))
	}

	// delete orphaned docker registry configmaps
	for _, dockerRegistryConfigMap := range dockerRegistryConfigMaps.Items {
		err := deleteOrphanedResource(ctx, m.GetClient(), &dockerRegistryConfigMap)
		if err != nil && !errors.IsNotFound(err) {
			collectedErrors = append(collectedErrors, fmt.Sprintf("failed to delete orphaned docker registry configmap %s/%s: %s", dockerRegistryConfigMap.Namespace, dockerRegistryConfigMap.Name, err))
		}
	}

	// list orphaned secrets
	secrets := &corev1.SecretList{}
	err = listOrphanedResources(ctx, m.GetAPIReader(), secrets, credentialsLabels)
	if err != nil {
		collectedErrors = append(collectedErrors, fmt.Sprintf("failed to list orphaned secrets: %s", err))
	}

	// delete orphaned secrets
	for _, secret := range secrets.Items {
		err := deleteOrphanedResource(ctx, m.GetClient(), &secret)
		if err != nil && !errors.IsNotFound(err) {
			collectedErrors = append(collectedErrors, fmt.Sprintf("failed to delete orphaned secret %s/%s: %s", secret.Namespace, secret.Name, err))
		}
	}

	// list orphaned service accounts
	serviceAccounts := &corev1.ServiceAccountList{}
	err = listServiceAccountsByName(ctx, m.GetAPIReader(), serviceAccounts, serviceAccountsName)
	if err != nil {
		collectedErrors = append(collectedErrors, fmt.Sprintf("failed to list orphaned service accounts: %s", err))
	}

	// delete orphaned service accounts
	for _, serviceAccount := range serviceAccounts.Items {
		err := deleteOrphanedResource(ctx, m.GetClient(), &serviceAccount)
		if err != nil && !errors.IsNotFound(err) {
			collectedErrors = append(collectedErrors, fmt.Sprintf("failed to delete orphaned service account %s/%s: %s", serviceAccount.Namespace, serviceAccount.Name, err))
		}
	}

	// remove BuildReady condition from Function status
	err = removeConditionFroStatus(ctx, m)
	if err != nil {
		collectedErrors = append(collectedErrors, fmt.Sprintf("failed to remove BuildReady condition from Function status: %s", err))
	}

	if len(collectedErrors) > 0 {
		return fmt.Errorf("orphaned resources collectedErrors:\n- %s", strings.Join(collectedErrors, "\n- "))
	}

	return nil
}

func listServiceAccountsByName(ctx context.Context, m client.Reader, resourceList *corev1.ServiceAccountList, name string) error {
	return m.List(ctx, resourceList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("metadata.name", name),
	})
}

func listOrphanedResources(ctx context.Context, m client.Reader, resourceList client.ObjectList, labels map[string]string) error {
	return m.List(ctx, resourceList, &client.ListOptions{
		LabelSelector: apilabels.SelectorFromSet(labels),
	})
}

func deleteOrphanedResource(ctx context.Context, m client.Client, resource client.Object) error {
	err := removeMatchingFinalizers(ctx, m, resource, "serverless.kyma-project.io/")
	if err != nil {
		return err
	}

	return m.Delete(ctx, resource, &client.DeleteOptions{
		PropagationPolicy: ptr.To(metav1.DeletePropagationBackground),
	})
}

func removeMatchingFinalizers(ctx context.Context, m client.Client, resource client.Object, prefix string) error {
	finalizers := resource.GetFinalizers()
	isFinalizerRemoved := false

	if len(finalizers) == 0 {
		return nil
	}

	for _, finalizer := range finalizers {
		if strings.HasPrefix(finalizer, prefix) {
			isFinalizerRemoved = isFinalizerRemoved || controllerutil.RemoveFinalizer(resource, finalizer)
		}
	}

	if isFinalizerRemoved {
		if err := m.Update(ctx, resource); err != nil {
			return fmt.Errorf("failed to remove finalizers from %s/%s: %s", resource.GetNamespace(), resource.GetName(), err)
		}
	}

	return nil
}

func removeConditionFroStatus(ctx context.Context, m manager.Manager) error {
	const (
		labelKey                = "serverless.kyma-project.io/managed-by"
		labelValue              = "function-controller"
		buildReadyConditionType = "BuildReady"
		functionGroup           = "serverless.kyma-project.io"
		functionVersion         = "v1alpha2"
		functionListKind        = "FunctionList"
		statusConditionsPathKey = "conditions"
	)

	fnList := &unstructured.UnstructuredList{}
	fnList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   functionGroup,
		Version: functionVersion,
		Kind:    functionListKind,
	})

	if err := m.GetClient().List(ctx, fnList, &client.ListOptions{
		LabelSelector: apilabels.SelectorFromSet(map[string]string{labelKey: labelValue}),
	}); err != nil {
		return fmt.Errorf("failed to list Functions: %w", err)
	}

	var collectedErrors []string

	for i := range fnList.Items {
		fn := &fnList.Items[i]

		conditions, found, err := unstructured.NestedSlice(fn.Object, "status", statusConditionsPathKey)
		if err != nil {
			collectedErrors = append(collectedErrors, fmt.Sprintf("failed to read status.conditions for %s/%s: %v", fn.GetNamespace(), fn.GetName(), err))
			continue
		}
		if !found || len(conditions) == 0 {
			continue
		}

		newConditions := make([]interface{}, 0, len(conditions))
		removed := false
		for _, c := range conditions {
			cMap, ok := c.(map[string]interface{})
			if !ok {
				newConditions = append(newConditions, c)
				continue
			}
			if t, ok := cMap["type"].(string); ok && t == buildReadyConditionType {
				removed = true
				continue
			}
			newConditions = append(newConditions, cMap)
		}

		if !removed {
			continue
		}

		if err := unstructured.SetNestedSlice(fn.Object, newConditions, "status", statusConditionsPathKey); err != nil {
			collectedErrors = append(collectedErrors, fmt.Sprintf("failed to set updated conditions for %s/%s: %v", fn.GetNamespace(), fn.GetName(), err))
			continue
		}

		if err := m.GetClient().Status().Update(ctx, fn); err != nil {
			collectedErrors = append(collectedErrors, fmt.Sprintf("failed to update Function status %s/%s: %v", fn.GetNamespace(), fn.GetName(), err))
			continue
		}
		m.GetLogger().Info("removed BuildReady condition", "namespace", fn.GetNamespace(), "name", fn.GetName())
	}

	if len(collectedErrors) > 0 {
		return fmt.Errorf("errors while removing BuildReady condition:\n- %s", strings.Join(collectedErrors, "\n- "))
	}

	return nil
}
