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
			serverlessDeploymentName = "serverless-ctrl-mngr"
			serverlessRegistrySecret = "serverless-registry-config-default"
			specSecretName           = "spec-secret-name"
		)

		var (
			dockerRegistryDataDefault = dockerRegistryData{
				TraceCollectorURL: ptr.To[string](v1alpha1.EndpointDisabled),
				EnableInternal:    ptr.To[bool](v1alpha1.DefaultEnableInternal),
				registrySecretData: registrySecretData{
					ServerAddress:   ptr.To[string](v1alpha1.DefaultServerAddress),
					RegistryAddress: ptr.To[string](v1alpha1.DefaultRegistryAddress),
				},
			}
			dockerRegistryDataWithChangedDependencies = dockerRegistryData{
				EventPublisherProxyURL: ptr.To[string]("test-eventing-address"),
				TraceCollectorURL:      ptr.To[string]("test-tracing-address"),
				EnableInternal:         ptr.To[bool](v1alpha1.DefaultEnableInternal),
				registrySecretData: registrySecretData{
					ServerAddress:   ptr.To[string](v1alpha1.DefaultServerAddress),
					RegistryAddress: ptr.To[string](v1alpha1.DefaultRegistryAddress),
				},
			}
			dockerRegistryDataExternalWithSecret = dockerRegistryData{
				EnableInternal: ptr.To[bool](false),
				registrySecretData: registrySecretData{
					Username:        ptr.To[string]("rumburak"),
					Password:        ptr.To[string]("mlekota"),
					ServerAddress:   ptr.To[string]("testserveraddress:5000"),
					RegistryAddress: ptr.To[string]("testregistryaddress:5000"),
				},
			}
			dockerRegistryDataExternalWithIncompleteSecret = dockerRegistryData{
				EnableInternal: ptr.To[bool](false),
				registrySecretData: registrySecretData{
					Username:      ptr.To[string]("blekota"),
					ServerAddress: ptr.To[string]("testserveraddress:5002"),
				},
			}
			dockerRegistryDataIncompleteFilledByDefault = dockerRegistryData{
				EnableInternal: ptr.To[bool](v1alpha1.DefaultEnableInternal),
				registrySecretData: registrySecretData{
					Username:        ptr.To[string]("blekota"),
					Password:        ptr.To[string](""),
					ServerAddress:   ptr.To[string]("testserveraddress:5002"),
					RegistryAddress: ptr.To[string](""),
				},
			}
			dockerRegistryDataExternalWithoutSecret = dockerRegistryData{
				EnableInternal: ptr.To[bool](false),
			}
		)

		It("The status should be Success", func() {
			h := testHelper{
				ctx:           context.Background(),
				namespaceName: namespaceName,
			}
			// TODO: implement test for enableInternal: true

			h.createNamespace()

			{
				emptyData := v1alpha1.DockerRegistrySpec{}
				shouldCreateServerless(h, serverlessName, serverlessDeploymentName, emptyData)
				shouldPropagateSpecProperties(h, serverlessDeploymentName, serverlessRegistrySecret, dockerRegistryDataDefault)
			}
			{
				updateData := v1alpha1.DockerRegistrySpec{
					Eventing: getEndpoint(dockerRegistryDataWithChangedDependencies.EventPublisherProxyURL),
					Tracing:  getEndpoint(dockerRegistryDataWithChangedDependencies.TraceCollectorURL),
				}
				shouldUpdateServerless(h, serverlessName, updateData)
				shouldPropagateSpecProperties(h, serverlessDeploymentName, serverlessRegistrySecret, dockerRegistryDataWithChangedDependencies)
			}
			{
				registryData := dockerRegistryDataExternalWithSecret
				secretName := specSecretName + "-full"
				h.createRegistrySecret(secretName, registryData.registrySecretData)
				updateData := registryData.toServerlessSpec(secretName)
				shouldUpdateServerless(h, serverlessName, updateData)
				shouldPropagateSpecProperties(h, serverlessDeploymentName, serverlessRegistrySecret, registryData)
			}
			{
				registryData := dockerRegistryDataExternalWithIncompleteSecret
				secretName := specSecretName + "-incomplete"
				h.createRegistrySecret(secretName, registryData.registrySecretData)
				updateData := registryData.toServerlessSpec(secretName)
				shouldUpdateServerless(h, serverlessName, updateData)
				shouldPropagateSpecProperties(h, serverlessDeploymentName, serverlessRegistrySecret, dockerRegistryDataIncompleteFilledByDefault)
			}
			{
				registryData := dockerRegistryDataExternalWithoutSecret
				updateData := registryData.toServerlessSpec("")
				shouldUpdateServerless(h, serverlessName, updateData)
				shouldPropagateSpecProperties(h, serverlessDeploymentName, serverlessRegistrySecret, dockerRegistryDataDefault)
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

func shouldPropagateSpecProperties(h testHelper, deploymentName, registrySecretName string, expected dockerRegistryData) {
	Eventually(h.createCheckRegistrySecretFunc(registrySecretName, expected.registrySecretData)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(BeTrue())

	Eventually(h.createCheckOptionalDependenciesFunc(deploymentName, expected)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(BeTrue())
}

func shouldUpdateServerless(h testHelper, serverlessName string, serverlessSpec v1alpha1.DockerRegistrySpec) {
	// arrange
	var serverless v1alpha1.DockerRegistry
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

	Eventually(h.createGetServerlessStatusFunc(serverlessName)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 20).
		Should(ConditionTrueMatcher())
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
