package controllers

import (
	"context"
	"fmt"
	"time"

	rtypes "github.com/kyma-project/module-manager/pkg/types"
	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type testHelper struct {
	ctx           context.Context
	namespaceName string
}

func (h *testHelper) updateDeploymentStatus(deploymentName string) {
	By(fmt.Sprintf("Updating deployment status: %s", deploymentName))
	var deployment appsv1.Deployment
	Eventually(h.createGetKubernetesObjectFunc(deploymentName, &deployment)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 30).
		Should(BeTrue())

	deployment.Status.Conditions = append(deployment.Status.Conditions, appsv1.DeploymentCondition{
		Type:    appsv1.DeploymentAvailable,
		Status:  corev1.ConditionTrue,
		Reason:  "test-reason",
		Message: "test-message",
	})
	deployment.Status.Replicas = 1
	Expect(k8sClient.Status().Update(h.ctx, &deployment)).To(Succeed())

	replicaSetName := h.createReplicaSetForDeployment(deployment)

	var replicaSet appsv1.ReplicaSet
	Eventually(h.createGetKubernetesObjectFunc(replicaSetName, &replicaSet)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 30).
		Should(BeTrue())

	replicaSet.Status.ReadyReplicas = 1
	replicaSet.Status.Replicas = 1
	Expect(k8sClient.Status().Update(h.ctx, &replicaSet)).To(Succeed())

	By(fmt.Sprintf("Deployment status updated: %s", deploymentName))
}

func (h *testHelper) createReplicaSetForDeployment(deployment appsv1.Deployment) string {
	replicaSetName := fmt.Sprintf("%s-replica-set", deployment.Name)
	By(fmt.Sprintf("Creating replica set (for deployment): %s", replicaSetName))
	var (
		trueValue = true
		one       = int32(1)
	)
	replicaSet := appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      replicaSetName,
			Namespace: h.namespaceName,
			Labels:    deployment.Spec.Selector.MatchLabels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       deployment.Name,
					UID:        deployment.GetUID(),
					Controller: &trueValue,
				},
			},
		},
		// dummy values
		Spec: appsv1.ReplicaSetSpec{
			Replicas: &one,
			Selector: deployment.Spec.Selector,
			Template: deployment.Spec.Template,
		},
	}
	Expect(k8sClient.Create(h.ctx, &replicaSet)).To(Succeed())
	By(fmt.Sprintf("Replica set (for deployment) created: %s", replicaSetName))
	return replicaSetName
}

func (h *testHelper) createServerless(serverlessName string, spec v1alpha1.ServerlessSpec) {
	By(fmt.Sprintf("Creating crd: %s", serverlessName))
	serverless := v1alpha1.Serverless{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serverlessName,
			Namespace: h.namespaceName,
			Labels: map[string]string{
				"operator.kyma-project.io/kyma-name": "test",
			},
		},
		Spec: spec,
	}
	Expect(k8sClient.Create(h.ctx, &serverless)).To(Succeed())
	By(fmt.Sprintf("Crd created: %s", serverlessName))
}

func (h *testHelper) createNamespace() {
	By(fmt.Sprintf("Creating namespace: %s", h.namespaceName))
	namespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: h.namespaceName,
		},
	}
	Expect(k8sClient.Create(h.ctx, &namespace)).To(Succeed())
	By(fmt.Sprintf("Namespace created: %s", h.namespaceName))
}

func (h *testHelper) createGetKubernetesObjectFunc(objectName string, obj client.Object) func() (bool, error) {
	return func() (bool, error) {
		return h.getKubernetesObjectFunc(objectName, obj)
	}
}

func (h *testHelper) getKubernetesObjectFunc(objectName string, obj client.Object) (bool, error) {
	key := types.NamespacedName{
		Name:      objectName,
		Namespace: h.namespaceName,
	}

	err := k8sClient.Get(h.ctx, key, obj)
	if err != nil {
		return false, err
	}
	return true, err
}

func (h *testHelper) createListKubernetesObjectFunc(list client.ObjectList) func() (bool, error) {
	return func() (bool, error) {
		return h.listKubernetesObjectFunc(list)
	}
}

func (h *testHelper) listKubernetesObjectFunc(list client.ObjectList) (bool, error) {
	opts := client.ListOptions{
		Namespace: h.namespaceName,
	}

	err := k8sClient.List(h.ctx, list, &opts)
	if err != nil {
		return false, err
	}
	return true, err
}

func (h *testHelper) createGetServerlessStateFunc(serverlessName string) func() (rtypes.State, error) {
	return func() (rtypes.State, error) {
		return h.getServerlessState(serverlessName)
	}
}

func (h *testHelper) getServerlessState(serverlessName string) (rtypes.State, error) {
	var emptyState = rtypes.State("")
	var serverless v1alpha1.Serverless
	key := types.NamespacedName{
		Name:      serverlessName,
		Namespace: h.namespaceName,
	}
	err := k8sClient.Get(h.ctx, key, &serverless)
	if err != nil {
		return emptyState, err
	}
	return serverless.Status.State, nil
}

func (h *testHelper) createCheckRegistrySecretFunc(serverlessRegistrySecret string, serverlessSpec v1alpha1.ServerlessSpec) func() (bool, error) {
	return func() (bool, error) {
		var configurationSecret corev1.Secret
		ok, err := h.getKubernetesObjectFunc(serverlessRegistrySecret, &configurationSecret)
		if !ok || err != nil {
			return ok, err
		}

		specRegistryAddress := *serverlessSpec.DockerRegistry.RegistryAddress
		registryAddress := string(configurationSecret.Data["registryAddress"])
		if registryAddress != specRegistryAddress {
			return false, fmt.Errorf("values not propagated( %s != %s )", registryAddress, specRegistryAddress)
		}

		specServerAddress := *serverlessSpec.DockerRegistry.ServerAddress
		serverAddress := string(configurationSecret.Data["serverAddress"])
		if serverAddress != specServerAddress {
			return false, fmt.Errorf("values not propagated( %s != %s )", registryAddress, specRegistryAddress)
		}

		return true, nil
	}
}
