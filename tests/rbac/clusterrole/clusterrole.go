package clusterrole

import (
	"context"
	"fmt"
	"time"

	serverlessv1alpha1 "github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	serverlessv1alpha2 "github.com/kyma-project/serverless/components/serverless/pkg/apis/serverless/v1alpha2"
	"github.com/kyma-project/serverless/tests/operator/utils"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// VerifyCRUD checks that the "edit" ClusterRole grants full CRUD on both
// functions.serverless.kyma-project.io and serverlesses.operator.kyma-project.io.
func VerifyCRUD(ctx context.Context, c client.Client, ns string, log *zap.SugaredLogger) error {
	log.Infof("verifying functions CRUD in namespace '%s'", ns)
	fn := &serverlessv1alpha2.Function{
		ObjectMeta: metav1.ObjectMeta{Name: "rbac-test-fn", Namespace: ns},
		Spec: serverlessv1alpha2.FunctionSpec{
			Runtime: serverlessv1alpha2.NodeJs22,
			Source:  serverlessv1alpha2.Source{Inline: &serverlessv1alpha2.InlineSource{Source: `module.exports = { main: function(e,c){ return "hello" } }`}},
		},
	}
	if err := crudObject(ctx, c, fn, &serverlessv1alpha2.FunctionList{}, ns); err != nil {
		return fmt.Errorf("functions CRUD failed: %w", err)
	}
	log.Info("functions CRUD succeeded ✓")

	log.Infof("verifying serverless CRUD in namespace '%s'", ns)
	svls := &serverlessv1alpha1.Serverless{
		ObjectMeta: metav1.ObjectMeta{Name: "rbac-test-svls", Namespace: ns},
		Spec:       serverlessv1alpha1.ServerlessSpec{DockerRegistry: &serverlessv1alpha1.DockerRegistry{EnableInternal: utils.PtrFromVal(false)}},
	}
	if err := crudObject(ctx, c, svls, &serverlessv1alpha1.ServerlessList{}, ns); err != nil {
		return fmt.Errorf("serverless CRUD failed: %w", err)
	}
	log.Info("serverless CRUD succeeded ✓")

	return nil
}

// VerifyCannotCreateNamespace checks that the "edit" ClusterRole does not allow creating cluster-scoped resources such as namespaces.
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
func crudObject(ctx context.Context, c client.Client, obj client.Object, list client.ObjectList, ns string) error {
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
func deleteIfExists(ctx context.Context, c client.Client, key client.ObjectKey, obj client.Object) {
	existing := obj.DeepCopyObject().(client.Object)
	if err := c.Get(ctx, key, existing); err == nil {
		_ = c.Delete(ctx, existing)
		time.Sleep(500 * time.Millisecond)
	}
}

// updateWithRetry re-fetches and updates the object, retrying on conflict.
func updateWithRetry(ctx context.Context, c client.Client, key client.ObjectKey, obj client.Object) error {
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
