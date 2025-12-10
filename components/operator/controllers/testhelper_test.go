package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/components/operator/internal/chart"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type conditionMatcher struct {
	expectedState           v1alpha1.State
	expectedConditionStatus metav1.ConditionStatus
}

func ConditionTrueMatcher() gomegatypes.GomegaMatcher {
	return &conditionMatcher{
		expectedState:           v1alpha1.StateReady,
		expectedConditionStatus: metav1.ConditionTrue,
	}
}

func (matcher *conditionMatcher) Match(actual interface{}) (success bool, err error) {
	status, ok := actual.(v1alpha1.ServerlessStatus)
	if !ok {
		return false, fmt.Errorf("ConditionMatcher matcher expects an v1alpha1.ServerlessStatus")
	}

	if status.State != matcher.expectedState {
		return false, nil
	}

	for _, condition := range status.Conditions {
		if condition.Status != matcher.expectedConditionStatus {
			return false, nil
		}
	}

	return true, nil
}

func (matcher *conditionMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto be in %s state with all %s conditions",
		actual, matcher.expectedState, matcher.expectedConditionStatus)
}

func (matcher *conditionMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto be in %s state with all %s conditions",
		actual, matcher.expectedState, matcher.expectedConditionStatus)
}

type testHelper struct {
	ctx           context.Context
	namespaceName string
}

func (h *testHelper) updateReplicaSetStatus(deploymentName string) {
	replicaSetName := fmt.Sprintf("%s-replica-set", deploymentName)

	By(fmt.Sprintf("Updating ReplicaSet status: %s", replicaSetName))

	var deployment appsv1.Deployment
	Eventually(h.createGetKubernetesObjectFunc(deploymentName, &deployment)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 30).
		Should(BeTrue())

	h.createReplicaSetForDeployment(replicaSetName, deployment)

	var replicaSet appsv1.ReplicaSet
	Eventually(h.createGetKubernetesObjectFunc(replicaSetName, &replicaSet)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 30).
		Should(BeTrue())

	replicaSet.Status.ReadyReplicas = 1
	replicaSet.Status.Replicas = 1
	Expect(k8sClient.Status().Update(h.ctx, &replicaSet)).To(Succeed())

	By(fmt.Sprintf("ReplicaSet status updated: %s", replicaSetName))
}

func (h *testHelper) updateDeploymentStatus(deploymentName string) {
	By(fmt.Sprintf("Updating deployment status: %s", deploymentName))
	var deployment appsv1.Deployment
	Eventually(h.createGetKubernetesObjectFunc(deploymentName, &deployment)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 30).
		Should(BeTrue())

	deployment.Status.Conditions = []appsv1.DeploymentCondition{
		{
			Type:    appsv1.DeploymentAvailable,
			Status:  corev1.ConditionTrue,
			Reason:  chart.MinimumReplicasAvailable,
			Message: "test-message",
		},
		{
			Type:    appsv1.DeploymentProgressing,
			Status:  corev1.ConditionTrue,
			Reason:  chart.NewRSAvailableReason,
			Message: "test-message",
		},
	}
	deployment.Status.ObservedGeneration = deployment.Generation
	deployment.Status.UnavailableReplicas = 0
	Expect(k8sClient.Status().Update(h.ctx, &deployment)).To(Succeed())

	By(fmt.Sprintf("Deployment status updated: %s", deploymentName))
}

func (h *testHelper) createReplicaSetForDeployment(replicaSetName string, deployment appsv1.Deployment) {
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
}

func (h *testHelper) createServerless(serverlessName string, spec v1alpha1.ServerlessSpec) {
	By(fmt.Sprintf("Creating crd: %s", serverlessName))
	serverless := v1alpha1.Serverless{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serverlessName,
			Namespace: h.namespaceName,
			//Annotations: map[string]string{
			//	"serverless.kyma-project.io/buildless-mode": "disabled",
			//},
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

func (h *testHelper) createSecret(name string, data map[string]string) {
	By(fmt.Sprintf("Creating secret: %s/%s", h.namespaceName, name))
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: h.namespaceName,
		},
		StringData: data,
	}
	Expect(k8sClient.Create(h.ctx, &secret)).To(Succeed())
	By(fmt.Sprintf("Secret created: %s/%s", h.namespaceName, name))
}

func (h *testHelper) createRegistrySecret(name string, data registrySecretData) {
	secretData := data.toMap()
	h.createSecret(name, secretData)
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

func (h *testHelper) createGetServerlessStatusFunc(serverlessName, deploymentName string) func() (v1alpha1.ServerlessStatus, error) {
	return func() (v1alpha1.ServerlessStatus, error) {
		// update deployment status to make sure .status.observedGeneration is up to date
		h.updateDeploymentStatus(deploymentName)
		return h.getServerlessStatus(serverlessName)
	}
}

func (h *testHelper) getServerlessStatus(serverlessName string) (v1alpha1.ServerlessStatus, error) {
	var serverless v1alpha1.Serverless
	key := types.NamespacedName{
		Name:      serverlessName,
		Namespace: h.namespaceName,
	}
	err := k8sClient.Get(h.ctx, key, &serverless)
	if err != nil {
		return v1alpha1.ServerlessStatus{}, err
	}
	return serverless.Status, nil
}

type serverlessData struct {
	EventPublisherProxyURL *string
	TraceCollectorURL      *string
	EnableInternal         *bool
	registrySecretData
}

func (d *serverlessData) toServerlessSpec(secretName string) v1alpha1.ServerlessSpec {
	result := v1alpha1.ServerlessSpec{
		Eventing: getEndpoint(d.EventPublisherProxyURL),
		Tracing:  getEndpoint(d.TraceCollectorURL),
		DockerRegistry: &v1alpha1.DockerRegistry{
			EnableInternal: d.EnableInternal,
		},
	}
	if secretName != "" {
		result.DockerRegistry.SecretName = ptr.To[string](secretName)
	}
	return result
}

func getEndpoint(url *string) *v1alpha1.Endpoint {
	if url != nil {
		return &v1alpha1.Endpoint{Endpoint: *url}
	}
	return nil
}

type registrySecretData struct {
	Username        *string
	Password        *string
	ServerAddress   *string
	RegistryAddress *string
}

func (d *registrySecretData) toMap() map[string]string {
	result := map[string]string{}
	if d.Username != nil {
		result["username"] = *d.Username
	}
	if d.Password != nil {
		result["password"] = *d.Password
	}
	if d.ServerAddress != nil {
		result["serverAddress"] = *d.ServerAddress
	}
	if d.RegistryAddress != nil {
		result["registryAddress"] = *d.RegistryAddress
	}
	return result
}

func (h *testHelper) createCheckRegistrySecretFunc(serverlessRegistrySecret string, expected registrySecretData) func() (bool, error) {
	return func() (bool, error) {
		var configurationSecret corev1.Secret

		if ok, err := h.getKubernetesObjectFunc(
			serverlessRegistrySecret, &configurationSecret); !ok || err != nil {
			return ok, err
		}
		if err := secretContainsSameValues(
			expected.toMap(), configurationSecret); err != nil {
			return false, err
		}
		if err := secretContainsRequired(configurationSecret); err != nil {
			return false, err
		}
		return true, nil
	}
}

func (h *testHelper) createCheckOptionalDependenciesFunc(deploymentName string, expected serverlessData) func() (bool, error) {
	return func() (bool, error) {
		var deploy appsv1.Deployment
		ok, err := h.getKubernetesObjectFunc(deploymentName, &deploy)
		if !ok || err != nil {
			return ok, err
		}

		eventProxyURL := v1alpha1.DefaultEventingEndpoint
		if expected.EventPublisherProxyURL != nil {
			eventProxyURL = *expected.EventPublisherProxyURL
		}

		traceCollectorURL := v1alpha1.EndpointDisabled
		if expected.TraceCollectorURL != nil {
			traceCollectorURL = *expected.TraceCollectorURL
		}

		if err := deploymentContainsEnv(deploy, "APP_FUNCTION_PUBLISHER_PROXY_ADDRESS", eventProxyURL); err != nil {
			return false, err
		}

		if err := deploymentContainsEnv(deploy, "APP_FUNCTION_TRACE_COLLECTOR_ENDPOINT", traceCollectorURL); err != nil {
			return false, err
		}

		return true, nil
	}
}

func deploymentContainsEnv(deployment appsv1.Deployment, name, value string) error {
	envs := deployment.Spec.Template.Spec.Containers[0].Env
	for i := range envs {
		if envs[i].Name == name && envs[i].Value == value {
			return nil
		}

		if envs[i].Name == name && envs[i].Value != value {
			return fmt.Errorf("wrong value for %s env: expected %s, got %s", name, value, envs[i].Value)
		}
	}

	return fmt.Errorf("env %s does not exist", name)
}

func secretContainsRequired(configurationSecret corev1.Secret) error {
	for _, k := range []string{"username", "password", "registryAddress", "serverAddress"} {
		_, ok := configurationSecret.Data[k]
		if !ok {
			return fmt.Errorf("values not propagated (%s is required)", k)
		}
	}
	return nil
}

func secretContainsSameValues(expected map[string]string, configurationSecret corev1.Secret) error {
	for k, expectedV := range expected {
		v, okV := configurationSecret.Data[k]
		if okV == false {
			return fmt.Errorf("values not propagated (%s: nil != %s )", k, expectedV)
		}
		if expectedV != string(v) {
			return fmt.Errorf("values not propagated (%s: %s != %s )", k, string(v), expectedV)
		}
	}
	return nil
}
