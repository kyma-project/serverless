package controllers

import (
	"context"
	"time"

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
			serverlessName           = "serverless-cr-test"
			serverlessDeploymentName = "serverless-ctrl-mngr"
			serverlessWebhookName    = "serverless-webhook-svc"
			serverlessRegistrySecret = "serverless-registry-config-default"
			specSecretName           = "spec-secret-name"
		)

		var (
			serverlessDataDefault = serverlessData{
				EventPublisherProxyURL: pointer.String(v1alpha1.DefaultPublisherProxyURL),
				TraceCollectorURL:      pointer.String(v1alpha1.DefaultTraceCollectorURL),
				EnableInternal:         pointer.Bool(v1alpha1.DefaultEnableInternal),
				registrySecretData: registrySecretData{
					ServerAddress:   pointer.String(v1alpha1.DefaultServerAddress),
					RegistryAddress: pointer.String(v1alpha1.DefaultRegistryAddress),
				},
			}
			serverlessDataWithChangedDependencies = serverlessData{
				EventPublisherProxyURL: pointer.String("test-address"),
				TraceCollectorURL:      pointer.String(""),
				EnableInternal:         pointer.Bool(v1alpha1.DefaultEnableInternal),
				registrySecretData: registrySecretData{
					ServerAddress:   pointer.String(v1alpha1.DefaultServerAddress),
					RegistryAddress: pointer.String(v1alpha1.DefaultRegistryAddress),
				},
			}
			serverlessDataExternalWithSecret = serverlessData{
				EnableInternal: pointer.Bool(false),
				registrySecretData: registrySecretData{
					Username:        pointer.String("rumburak"),
					Password:        pointer.String("mlekota"),
					ServerAddress:   pointer.String("testserveraddress:5000"),
					RegistryAddress: pointer.String("testregistryaddress:5000"),
				},
			}
			serverlessDataExternalWithIncompleteSecret = serverlessData{
				EnableInternal: pointer.Bool(false),
				registrySecretData: registrySecretData{
					Username:      pointer.String("blekota"),
					ServerAddress: pointer.String("testserveraddress:5002"),
				},
			}
			serverlessDataIncompleteFilledByDefault = serverlessData{
				EnableInternal: pointer.Bool(v1alpha1.DefaultEnableInternal),
				registrySecretData: registrySecretData{
					Username:        pointer.String("blekota"),
					Password:        pointer.String(""),
					ServerAddress:   pointer.String("testserveraddress:5002"),
					RegistryAddress: pointer.String(""),
				},
			}
			serverlessDataExternalWithoutSecret = serverlessData{
				EnableInternal: pointer.Bool(false),
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
				shouldCreateServerless(h, serverlessName, serverlessDeploymentName, serverlessWebhookName, emptyData)
				shouldPropagateSpecProperties(h, serverlessDeploymentName, serverlessRegistrySecret, serverlessDataDefault)
			}
			{
				updateData := v1alpha1.ServerlessSpec{
					Eventing: getEndpoint(serverlessDataWithChangedDependencies.EventPublisherProxyURL),
					Tracing:  getEndpoint(serverlessDataWithChangedDependencies.TraceCollectorURL),
				}
				shouldUpdateServerless(h, serverlessName, updateData)
				shouldPropagateSpecProperties(h, serverlessDeploymentName, serverlessRegistrySecret, serverlessDataWithChangedDependencies)
			}
			{
				registryData := serverlessDataExternalWithSecret
				secretName := specSecretName + "-full"
				h.createRegistrySecret(secretName, registryData.registrySecretData)
				updateData := registryData.toServerlessSpec(secretName)
				shouldUpdateServerless(h, serverlessName, updateData)
				shouldPropagateSpecProperties(h, serverlessDeploymentName, serverlessRegistrySecret, registryData)
			}
			{
				registryData := serverlessDataExternalWithIncompleteSecret
				secretName := specSecretName + "-incomplete"
				h.createRegistrySecret(secretName, registryData.registrySecretData)
				updateData := registryData.toServerlessSpec(secretName)
				shouldUpdateServerless(h, serverlessName, updateData)
				shouldPropagateSpecProperties(h, serverlessDeploymentName, serverlessRegistrySecret, serverlessDataIncompleteFilledByDefault)
			}
			{
				registryData := serverlessDataExternalWithoutSecret
				updateData := registryData.toServerlessSpec("")
				shouldUpdateServerless(h, serverlessName, updateData)
				shouldPropagateSpecProperties(h, serverlessDeploymentName, serverlessRegistrySecret, serverlessDataDefault)
			}

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
	Eventually(h.createGetServerlessStatusFunc(serverlessName)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 20).
		Should(ConditionTrueMatcher())
}

func shouldPropagateSpecProperties(h testHelper, deploymentName, registrySecretName string, expected serverlessData) {
	Eventually(h.createCheckRegistrySecretFunc(registrySecretName, expected.registrySecretData)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(BeTrue())

	Eventually(h.createCheckOptionalDependenciesFunc(deploymentName, expected)).
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

	Eventually(h.createGetServerlessStatusFunc(serverlessName)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 20).
		Should(ConditionTrueMatcher())
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

	Eventually(h.createGetKubernetesObjectFunc(serverlessName, &serverless)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(BeTrue())

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
