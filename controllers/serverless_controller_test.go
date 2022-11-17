package controllers

import (
	"context"
	"time"

	rtypes "github.com/kyma-project/module-manager/operator/pkg/types"
	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/utils/pointer"
)

var _ = Describe("Serverless controller", func() {
	Context("When creating fresh instance", func() {
		const (
			namespaceName            = "kyma-system"
			serverlessName           = "test"
			serverlessDeploymentName = "serverless-ctrl-mngr"
			serverlessWebhookName    = "serverless-webhook-svc"
			serverlessRegistrySecret = "serverless-registry-config-default"
			operatorName             = "keda-manager"
		)

		var (
			serverlessSpec = v1alpha1.ServerlessSpec{
				DockerRegistry: &v1alpha1.DockerRegistry{
					EnableInternal:  pointer.Bool(false),
					ServerAddress:   pointer.String("testaddress:5000"),
					RegistryAddress: pointer.String("testaddress:5000"),
				},
			}
			serverlessUpdatedSpec = v1alpha1.ServerlessSpec{
				DockerRegistry: &v1alpha1.DockerRegistry{
					EnableInternal:  pointer.Bool(false),
					ServerAddress:   pointer.String("othertestaddress:5000"),
					RegistryAddress: pointer.String("othertestaddress:5000"),
				},
			}
		)

		It("The status should be Success", func() {
			h := testHelper{
				ctx:           context.Background(),
				namespaceName: namespaceName,
			}
			h.createNamespace()

			shouldCreateServerless(h, serverlessName, serverlessDeploymentName, serverlessWebhookName, serverlessSpec)

			shouldPropagateSpecProperties(h, serverlessRegistrySecret, serverlessSpec)

			shouldUpdateServerless(h, serverlessName, serverlessUpdatedSpec)

			shouldPropagateSpecProperties(h, serverlessRegistrySecret, serverlessUpdatedSpec)

			shouldDeleteServerless(h, serverlessName, serverlessDeploymentName, serverlessWebhookName)
		})
	})
})

func shouldCreateServerless(h testHelper, serverlessName, serverlessDeploymentName, serverlessWebhookName string, spec v1alpha1.ServerlessSpec) {
	// act
	h.createServerless(serverlessName, spec)

	// we have to update deployment status manually
	h.updateDeploymentStatus(serverlessDeploymentName)
	h.updateDeploymentStatus(serverlessWebhookName)

	// assert
	Eventually(h.createGetServerlessStateFunc(serverlessName)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 20).
		Should(Equal(rtypes.StateReady))
}

func shouldPropagateSpecProperties(h testHelper, registrySecretName string, serverlessSpec v1alpha1.ServerlessSpec) {
	// TODO: implement more propagation checks here

	Eventually(h.createCheckRegistrySecretFunc(registrySecretName, serverlessSpec)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(BeTrue())
}

func shouldUpdateServerless(h testHelper, serverlessName string, serverlessSpec v1alpha1.ServerlessSpec) {
	// arrange
	var serverless v1alpha1.Serverless
	Eventually(h.createGetKubernetesObjectFunc(serverlessName, &serverless)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(BeTrue())

	serverless.Spec = serverlessSpec

	// act
	Expect(k8sClient.Update(h.ctx, &serverless)).To(Succeed())

	// assert
	Eventually(h.createGetKubernetesObjectFunc(serverlessName, &serverless)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(BeTrue())

	Expect(serverless.Spec).To(Equal(serverlessSpec))
}

func shouldDeleteServerless(h testHelper, serverlessName, serverlessDeploymentName, serverlessWebhookName string) {
	// initial assert
	var deployList appsv1.DeploymentList
	Eventually(h.createListKubernetesObjectFunc(&deployList)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(BeTrue())

	Expect(deployList.Items).To(HaveLen(2))

	// act
	var serverless v1alpha1.Serverless
	Eventually(h.createGetKubernetesObjectFunc(serverlessName, &serverless)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(BeTrue())

	Expect(k8sClient.Delete(h.ctx, &serverless)).To(Succeed())

	// assert
	Eventually(h.createGetKubernetesObjectFunc(serverlessDeploymentName, &appsv1.Deployment{})).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(BeTrue())

	Eventually(h.createGetKubernetesObjectFunc(serverlessWebhookName, &appsv1.Deployment{})).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(BeTrue())
}
