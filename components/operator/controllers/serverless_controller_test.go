package controllers

import (
	"context"
	"time"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("DockerRegistry controller", func() {
	Context("When creating fresh instance", func() {
		const (
			namespaceName            = "kyma-system"
			serverlessName           = "serverless-cr-test"
			serverlessDeploymentName = "internal-docker-registry"
			serverlessRegistrySecret = "serverless-registry-config-default"
		)

		var (
			dockerRegistryDataDefault = dockerRegistryData{
				TraceCollectorURL: ptr.To[string](v1alpha1.EndpointDisabled),
				EnableInternal:    ptr.To[bool](v1alpha1.DefaultEnableInternal),
			}
		)

		It("The status should be Success", func() {
			h := testHelper{
				ctx:           context.Background(),
				namespaceName: namespaceName,
			}
			h.createNamespace()

			{
				emptyData := v1alpha1.DockerRegistrySpec{
					DockerRegistry: &v1alpha1.DockerRegistryCfg{
						EnableInternal: ptr.To[bool](true),
					},
				}
				shouldCreateServerless(h, serverlessName, serverlessDeploymentName, emptyData)
				shouldPropagateSpecProperties(h, serverlessRegistrySecret, dockerRegistryDataDefault)
			}

			shouldDeleteServerless(h, serverlessName, serverlessDeploymentName)
		})
	})
})

func shouldCreateServerless(h testHelper, serverlessName, serverlessDeploymentName string, spec v1alpha1.DockerRegistrySpec) {
	// act
	h.createServerless(serverlessName, spec)

	// we have to update deployment status manually
	h.updateDeploymentStatus(serverlessDeploymentName)

	// assert
	Eventually(h.createGetServerlessStatusFunc(serverlessName)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 20).
		Should(ConditionTrueMatcher())
}

func shouldPropagateSpecProperties(h testHelper, registrySecretName string, expected dockerRegistryData) {
	Eventually(h.createCheckRegistrySecretFunc(registrySecretName, expected.registrySecretData)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(BeTrue())
}

func shouldDeleteServerless(h testHelper, serverlessName, serverlessDeploymentName string) {
	// initial assert
	var deployList appsv1.DeploymentList
	Eventually(h.createListKubernetesObjectFunc(&deployList)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(BeTrue())

	Expect(deployList.Items).To(HaveLen(1))

	// act
	var serverless v1alpha1.DockerRegistry
	Eventually(h.createGetKubernetesObjectFunc(serverlessName, &serverless)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(BeTrue())

	Expect(k8sClient.Delete(h.ctx, &serverless)).To(Succeed())

	Eventually(h.createGetKubernetesObjectFunc(serverlessName, &serverless)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(BeTrue())

	// assert
	Eventually(h.createGetKubernetesObjectFunc(serverlessDeploymentName, &appsv1.Deployment{})).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(BeTrue())
}
