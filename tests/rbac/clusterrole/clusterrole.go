package clusterrole

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/serverless/tests/rbac/resource"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// VerifyEditCRUD checks that the "edit" ClusterRole grants full CRUD on each of the specified resources.
func VerifyEditCRUD(ctx context.Context, c client.Client, ns string, resources []resource.TestResource, log *zap.SugaredLogger) error {
	for _, r := range resources {
		log.Infof("verifying edit CRUD for %s in namespace '%s'", r.Name(), ns)
		obj := r.Object(ns)
		list := r.ObjectList()
		if err := crudObject(ctx, c, obj, list, ns); err != nil {
			return fmt.Errorf("%s CRUD failed: %w", r.Name(), err)
		}
		log.Infof("%s CRUD succeeded ✓", r.Name())
	}
	return nil
}

// PreCreateResources creates sample objects for each resource so they exist for other scenarios.
func PreCreateResources(ctx context.Context, c client.Client, ns string, resources []resource.TestResource, log *zap.SugaredLogger) error {
	for _, r := range resources {
		obj := r.Object(ns)
		log.Infof("pre-creating %s '%s/%s'", r.Name(), obj.GetNamespace(), obj.GetName())
		key := client.ObjectKeyFromObject(obj)
		deleteIfExists(ctx, c, key, obj)
		if err := c.Create(ctx, obj); err != nil {
			return fmt.Errorf("pre-create %s: %w", r.Name(), err)
		}
	}
	return nil
}

// VerifyViewReadOnly checks that the "view" ClusterRole allows reading (get/list)
// but forbids create, update, and delete for each of the specified resources.
// The resources must already exist (use PreCreateResources first).
func VerifyViewReadOnly(ctx context.Context, c client.Client, ns string, resources []resource.TestResource, log *zap.SugaredLogger) error {
	for _, r := range resources {
		log.Infof("verifying view read-only for %s in namespace '%s'", r.Name(), ns)
		obj := r.Object(ns)
		list := r.ObjectList()
		key := client.ObjectKeyFromObject(obj)

		// GET should succeed (resource was pre-created)
		existing := obj.DeepCopy()
		if err := c.Get(ctx, key, existing); err != nil {
			return fmt.Errorf("%s get should succeed: %w", r.Name(), err)
		}
		log.Infof("%s get succeeded ✓", r.Name())

		// LIST should succeed
		if err := c.List(ctx, list, client.InNamespace(ns)); err != nil {
			return fmt.Errorf("%s list should succeed: %w", r.Name(), err)
		}
		log.Infof("%s list succeeded ✓", r.Name())

		// CREATE should be forbidden
		newObj := r.Object(ns)
		newObj.SetName(newObj.GetName() + "-view-test")
		if err := c.Create(ctx, newObj); err == nil {
			_ = c.Delete(ctx, newObj)
			return fmt.Errorf("%s create should be forbidden but succeeded", r.Name())
		} else if !k8serrors.IsForbidden(err) {
			return fmt.Errorf("%s create: expected Forbidden, got: %w", r.Name(), err)
		}
		log.Infof("%s create correctly forbidden ✓", r.Name())

		// UPDATE should be forbidden
		existing.SetLabels(map[string]string{"rbac-test": "should-fail"})
		if err := c.Update(ctx, existing); err == nil {
			return fmt.Errorf("%s update should be forbidden but succeeded", r.Name())
		} else if !k8serrors.IsForbidden(err) {
			return fmt.Errorf("%s update: expected Forbidden, got: %w", r.Name(), err)
		}
		log.Infof("%s update correctly forbidden ✓", r.Name())

		// DELETE should be forbidden
		if err := c.Delete(ctx, existing); err == nil {
			return fmt.Errorf("%s delete should be forbidden but succeeded", r.Name())
		} else if !k8serrors.IsForbidden(err) {
			return fmt.Errorf("%s delete: expected Forbidden, got: %w", r.Name(), err)
		}
		log.Infof("%s delete correctly forbidden ✓", r.Name())

		log.Infof("%s view read-only verified ✓", r.Name())
	}

	// Secrets must not be readable with the view ClusterRole
	log.Info("verifying secrets are not readable")
	secretList := &corev1.SecretList{}
	if err := c.List(ctx, secretList, client.InNamespace(ns)); err == nil {
		return fmt.Errorf("secrets list should be forbidden but succeeded")
	} else if !k8serrors.IsForbidden(err) {
		return fmt.Errorf("secrets list: expected Forbidden, got: %w", err)
	}
	log.Info("secrets list correctly forbidden ✓")

	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "rbac-test-secret", Namespace: ns}}
	if err := c.Get(ctx, client.ObjectKeyFromObject(secret), secret); err == nil {
		return fmt.Errorf("secrets get should be forbidden but succeeded")
	} else if !k8serrors.IsForbidden(err) {
		return fmt.Errorf("secrets get: expected Forbidden, got: %w", err)
	}
	log.Info("secrets get correctly forbidden ✓")

	return nil
}

// VerifyCannotCreateNamespace checks that the ClusterRole does not allow creating cluster-scoped resources such as namespaces.
func VerifyCannotCreateNamespace(ctx context.Context, c client.Client, log *zap.SugaredLogger) error {
	log.Info("verifying namespace creation is denied")
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "rbac-test-should-be-forbidden"}}
	err := c.Create(ctx, ns)
	if err == nil {
		_ = c.Delete(ctx, ns)
		return fmt.Errorf("expected Forbidden but namespace creation succeeded")
	}
	if !k8serrors.IsForbidden(err) {
		return fmt.Errorf("unexpected error (expected Forbidden): %w", err)
	}
	log.Info("namespace creation correctly forbidden ✓")
	return nil
}

// crudObject runs create → get → update → list → delete
func crudObject(ctx context.Context, c client.Client, obj *unstructured.Unstructured, list *unstructured.UnstructuredList, ns string) error {
	key := client.ObjectKeyFromObject(obj)
	deleteIfExists(ctx, c, key, obj)

	if err := c.Create(ctx, obj); err != nil {
		return fmt.Errorf("create: %w", err)
	}
	defer func() { _ = c.Delete(ctx, obj) }()

	if err := updateWithRetry(ctx, c, key, obj); err != nil {
		return err
	}

	if err := c.List(ctx, list, client.InNamespace(ns)); err != nil {
		return fmt.Errorf("list: %w", err)
	}

	if err := c.Delete(ctx, obj); err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

// deleteIfExists removes a leftover object from a previous failed run.
func deleteIfExists(ctx context.Context, c client.Client, key client.ObjectKey, obj *unstructured.Unstructured) {
	existing := obj.DeepCopy()
	if err := c.Get(ctx, key, existing); err == nil {
		_ = c.Delete(ctx, existing)
		time.Sleep(500 * time.Millisecond)
	}
}

// updateWithRetry re-fetches and updates the object, retrying on conflict.
func updateWithRetry(ctx context.Context, c client.Client, key client.ObjectKey, obj *unstructured.Unstructured) error {
	const maxAttempts = 5
	for i := range maxAttempts {
		if err := c.Get(ctx, key, obj); err != nil {
			return fmt.Errorf("get: %w", err)
		}
		labels := obj.GetLabels()
		if labels == nil {
			labels = map[string]string{}
		}
		labels["rbac-test"] = "ok"
		obj.SetLabels(labels)

		err := c.Update(ctx, obj)
		if err == nil {
			return nil
		}
		if !k8serrors.IsConflict(err) {
			return fmt.Errorf("update: %w", err)
		}
		if i < maxAttempts-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}
	return fmt.Errorf("update: failed after %d conflict retries", maxAttempts)
}
