package orphaned_resources

import (
	"context"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
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

	serviceAccountsLabels := map[string]string{
		"app.kubernetes.io/component": "serverless",
		"app.kubernetes.io/part-of":   "serverless",
	}

	operatorServiceAccountsLabels := map[string]string{
		"app.kubernetes.io/component": "serverless-rbac",
		"app.kubernetes.io/name":      "serverless-operator",
	}

	// list orphaned jobs
	jobs := &batchv1.JobList{}
	err := listOrphanedResources(ctx, m.GetAPIReader(), jobs, labels)
	if err != nil {
		return fmt.Errorf("failed to list orphaned jobs: %s", err)
	}

	// delete orphaned jobs
	for _, job := range jobs.Items {
		err := deleteOrphanedResource(ctx, m.GetClient(), &job)
		if err != nil {
			return fmt.Errorf("failed to delete orphaned-resources %s/%s: %s", job.Namespace, job.Name, err)
		}
	}

	// list orphaned configmaps
	configMaps := &corev1.ConfigMapList{}
	err = listOrphanedResources(ctx, m.GetAPIReader(), configMaps, labels)
	if err != nil {
		return fmt.Errorf("failed to list orphaned configmaps: %s", err)
	}

	// delete orphaned configmaps
	for _, configMap := range configMaps.Items {
		err := deleteOrphanedResource(ctx, m.GetClient(), &configMap)
		if err != nil {
			return fmt.Errorf("failed to delete orphaned configmap %s/%s: %s", configMap.Namespace, configMap.Name, err)
		}
	}

	// list orphaned runtime configmaps
	runtimeConfigMaps := &corev1.ConfigMapList{}
	err = listOrphanedResources(ctx, m.GetAPIReader(), runtimeConfigMaps, runtimeLabels)
	if err != nil {
		return fmt.Errorf("failed to list orphaned runtime configmaps: %s", err)
	}

	// delete orphaned runtime configmaps
	for _, runtimeConfigMap := range runtimeConfigMaps.Items {
		err := deleteOrphanedResource(ctx, m.GetClient(), &runtimeConfigMap)
		if err != nil {
			return fmt.Errorf("failed to delete orphaned runtime configmap %s/%s: %s", runtimeConfigMap.Namespace, runtimeConfigMap.Name, err)
		}
	}

	// list orphaned docker registry configmaps
	dockerRegistryConfigMaps := &corev1.ConfigMapList{}
	err = listOrphanedResources(ctx, m.GetAPIReader(), dockerRegistryConfigMaps, dockerRegistryLabels)
	if err != nil {
		return fmt.Errorf("failed to list orphaned docker registry configmaps: %s", err)
	}

	// delete orphaned docker registry configmaps
	for _, dockerRegistryConfigMap := range dockerRegistryConfigMaps.Items {
		err := deleteOrphanedResource(ctx, m.GetClient(), &dockerRegistryConfigMap)
		if err != nil {
			return fmt.Errorf("failed to delete orphaned docker registry configmap %s/%s: %s", dockerRegistryConfigMap.Namespace, dockerRegistryConfigMap.Name, err)
		}
	}

	// list orphaned secrets
	secrets := &corev1.SecretList{}
	err = listOrphanedResources(ctx, m.GetAPIReader(), secrets, credentialsLabels)
	if err != nil {
		return fmt.Errorf("failed to list orphaned secrets: %s", err)
	}

	// delete orphaned secrets
	for _, secret := range secrets.Items {
		err := deleteOrphanedResource(ctx, m.GetClient(), &secret)
		if err != nil {
			return fmt.Errorf("failed to delete orphaned secret %s/%s: %s", secret.Namespace, secret.Name, err)
		}
	}

	// list orphaned service accounts
	serviceAccounts := &corev1.ServiceAccountList{}
	err = listOrphanedResources(ctx, m.GetAPIReader(), serviceAccounts, serviceAccountsLabels)
	if err != nil {
		return fmt.Errorf("failed to list orphaned service accounts: %s", err)
	}

	// delete orphaned service accounts
	for _, serviceAccount := range serviceAccounts.Items {
		err := deleteOrphanedResource(ctx, m.GetClient(), &serviceAccount)
		if err != nil {
			return fmt.Errorf("failed to delete orphaned service account %s/%s: %s", serviceAccount.Namespace, serviceAccount.Name, err)
		}
	}

	// list orphaned operator service accounts
	operatorServiceAccounts := &corev1.ServiceAccountList{}
	err = listOrphanedResources(ctx, m.GetAPIReader(), operatorServiceAccounts, operatorServiceAccountsLabels)
	if err != nil {
		return fmt.Errorf("failed to list orphaned service accounts: %s", err)
	}

	// delete orphaned operator service accounts
	for _, operatorServiceAccount := range operatorServiceAccounts.Items {
		err := deleteOrphanedResource(ctx, m.GetClient(), &operatorServiceAccount)
		if err != nil {
			return fmt.Errorf("failed to delete orphaned service account %s/%s: %s", operatorServiceAccount.Namespace, operatorServiceAccount.Name, err)
		}
	}

	return nil
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
