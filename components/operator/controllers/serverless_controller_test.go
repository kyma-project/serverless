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

const (
	serverlessDeploymentName = "serverless-ctrl-mngr"
	serverlessName           = "serverless-cr-test"
	serverlessRegistrySecret = "serverless-registry-config-default"
)

var _ = Describe("Serverless controller", func() {
	Context("When creating fresh instance", func() {
		const (
			namespaceName  = "kyma-system"
			specSecretName = "spec-secret-name"
		)

		var (
			serverlessDataDefault = serverlessData{
				TraceCollectorURL: ptr.To[string](v1alpha1.EndpointDisabled),
				EnableInternal:    ptr.To[bool](v1alpha1.DefaultEnableInternal),
				registrySecretData: registrySecretData{
					ServerAddress:   ptr.To[string](v1alpha1.DefaultServerAddress),
					RegistryAddress: ptr.To[string](v1alpha1.DefaultRegistryAddress),
				},
			}
			serverlessDataWithChangedDependencies = serverlessData{
				EventPublisherProxyURL: ptr.To[string]("test-eventing-address"),
				TraceCollectorURL:      ptr.To[string]("test-tracing-address"),
				EnableInternal:         ptr.To[bool](v1alpha1.DefaultEnableInternal),
				registrySecretData: registrySecretData{
					ServerAddress:   ptr.To[string](v1alpha1.DefaultServerAddress),
					RegistryAddress: ptr.To[string](v1alpha1.DefaultRegistryAddress),
				},
			}
			serverlessDataExternalWithSecret = serverlessData{
				EnableInternal: ptr.To[bool](false),
				registrySecretData: registrySecretData{
					Username:        ptr.To[string]("rumburak"),
					Password:        ptr.To[string]("mlekota"),
					ServerAddress:   ptr.To[string]("testserveraddress:5000"),
					RegistryAddress: ptr.To[string]("testregistryaddress:5000"),
				},
			}
			serverlessDataExternalWithIncompleteSecret = serverlessData{
				EnableInternal: ptr.To[bool](false),
				registrySecretData: registrySecretData{
					Username:      ptr.To[string]("blekota"),
					ServerAddress: ptr.To[string]("testserveraddress:5002"),
				},
			}
			serverlessDataIncompleteFilledByDefault = serverlessData{
				EnableInternal: ptr.To[bool](v1alpha1.DefaultEnableInternal),
				registrySecretData: registrySecretData{
					Username:        ptr.To[string]("blekota"),
					Password:        ptr.To[string](""),
					ServerAddress:   ptr.To[string]("testserveraddress:5002"),
					RegistryAddress: ptr.To[string](""),
				},
			}
			serverlessDataExternalWithoutSecret = serverlessData{
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
				emptyData := v1alpha1.ServerlessSpec{}
				shouldCreateServerless(h, serverlessName, serverlessDeploymentName, emptyData)
				shouldPropagateSpecProperties(h, serverlessDataDefault)
			}
			{
				updateData := v1alpha1.ServerlessSpec{
					Eventing: getEndpoint(serverlessDataWithChangedDependencies.EventPublisherProxyURL),
					Tracing:  getEndpoint(serverlessDataWithChangedDependencies.TraceCollectorURL),
				}
				shouldUpdateServerless(h, updateData)
				shouldPropagateSpecProperties(h, serverlessDataWithChangedDependencies)
			}
			{
				registryData := serverlessDataExternalWithSecret
				secretName := specSecretName + "-full"
				h.createRegistrySecret(secretName, registryData.registrySecretData)
				updateData := registryData.toServerlessSpec(secretName)
				shouldUpdateServerless(h, updateData)
				shouldPropagateSpecProperties(h, registryData)
			}
			{
				registryData := serverlessDataExternalWithIncompleteSecret
				secretName := specSecretName + "-incomplete"
				h.createRegistrySecret(secretName, registryData.registrySecretData)
				updateData := registryData.toServerlessSpec(secretName)
				shouldUpdateServerless(h, updateData)
				shouldPropagateSpecProperties(h, serverlessDataIncompleteFilledByDefault)
			}
			{
				registryData := serverlessDataExternalWithoutSecret
				updateData := registryData.toServerlessSpec("")
				shouldUpdateServerless(h, updateData)
				shouldPropagateSpecProperties(h, serverlessDataDefault)
			}

			shouldDeleteServerless(h, serverlessName, serverlessDeploymentName)
		})
	})
})

func shouldCreateServerless(h testHelper, serverlessName, serverlessDeploymentName string, spec v1alpha1.ServerlessSpec) {
	// act
	h.createServerless(serverlessName, spec)

	// we have to update deployment status manually
	h.updateDeploymentStatus(serverlessDeploymentName)
	h.updateReplicaSetStatus(serverlessDeploymentName)

	// assert
	Eventually(h.createGetServerlessStatusFunc(serverlessName, serverlessDeploymentName)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 20).
		Should(ConditionTrueMatcher())
}

func shouldPropagateSpecProperties(h testHelper, expected serverlessData) {
	Eventually(h.createCheckOptionalDependenciesFunc(serverlessDeploymentName, expected)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(BeTrue())
}

func shouldUpdateServerless(h testHelper, serverlessSpec v1alpha1.ServerlessSpec) {
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

	Eventually(h.createGetServerlessStatusFunc(serverlessName, serverlessDeploymentName)).
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
	var serverless v1alpha1.Serverless
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
