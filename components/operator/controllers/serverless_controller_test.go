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
			namespaceName  = "kyma-system"
			crName         = "cr-test"
			deploymentName = "internal-docker-registry"
			registrySecret = "serverless-registry-config-default"
		)

		var (
			defaultData = dockerRegistryData{
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
				emptyData := v1alpha1.DockerRegistrySpec{}
				shouldCreateDockerRegistry(h, crName, deploymentName, emptyData)
				shouldPropagateSpecProperties(h, registrySecret, defaultData)
			}

			shouldDeleteDockerRegistry(h, crName, deploymentName)
		})
	})
})

func shouldCreateDockerRegistry(h testHelper, name, deploymentName string, spec v1alpha1.DockerRegistrySpec) {
	// act
	h.createDockerRegistry(name, spec)

	// we have to update deployment status manually
	h.updateDeploymentStatus(deploymentName)

	// assert
	Eventually(h.getDockerRegistryStatusFunc(name)).
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

func shouldDeleteDockerRegistry(h testHelper, name, deploymentName string) {
	// initial assert
	var deployList appsv1.DeploymentList
	Eventually(h.listKubernetesObjectFunc(&deployList)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(BeTrue())

	Expect(deployList.Items).To(HaveLen(1))

	// act
	var dockerRegistry v1alpha1.DockerRegistry
	Eventually(h.getKubernetesObjectFunc(name, &dockerRegistry)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(BeTrue())

	Expect(k8sClient.Delete(h.ctx, &dockerRegistry)).To(Succeed())

	Eventually(h.getKubernetesObjectFunc(name, &dockerRegistry)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(BeTrue())

	// assert
	Eventually(h.getKubernetesObjectFunc(deploymentName, &appsv1.Deployment{})).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(BeTrue())
}
