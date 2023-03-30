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
			registryDataDefault = dockerRegistryData{
				EnableInternal: pointer.Bool(v1alpha1.DefaultEnableInternal),
				registrySecretData: registrySecretData{
					ServerAddress:   pointer.String(v1alpha1.DefaultServerAddress),
					RegistryAddress: pointer.String(v1alpha1.DefaultRegistryAddress),
				},
			}
			registryDataExternalWithSecret = dockerRegistryData{
				EnableInternal: pointer.Bool(false),
				registrySecretData: registrySecretData{
					Username:        pointer.String("rumburak"),
					Password:        pointer.String("mlekota"),
					ServerAddress:   pointer.String("testserveraddress:5000"),
					RegistryAddress: pointer.String("testregistryaddress:5000"),
				},
			}
			registryDataExternalWithIncompleteSecret = dockerRegistryData{
				EnableInternal: pointer.Bool(false),
				registrySecretData: registrySecretData{
					Username:      pointer.String("blekota"),
					ServerAddress: pointer.String("testserveraddress:5002"),
				},
			}
			registryDataIncompleteFilledByDefault = dockerRegistryData{
				EnableInternal: pointer.Bool(v1alpha1.DefaultEnableInternal),
				registrySecretData: registrySecretData{
					Username:        pointer.String("blekota"),
					ServerAddress:   pointer.String("testserveraddress:5002"),
					RegistryAddress: pointer.String(v1alpha1.DefaultRegistryAddress),
				},
			}
			registryDataExternalWithoutSecret = dockerRegistryData{
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
				shouldPropagateSpecProperties(h, serverlessRegistrySecret, registryDataDefault)
			}
			{
				registryData := registryDataExternalWithSecret
				secretName := specSecretName + "-full"
				h.createRegistrySecret(secretName, registryData.registrySecretData)
				updateData := registryData.toServerlessSpec(secretName)
				shouldUpdateServerless(h, serverlessName, updateData)
				shouldPropagateSpecProperties(h, serverlessRegistrySecret, registryData)
			}
			{
				registryData := registryDataExternalWithIncompleteSecret
				secretName := specSecretName + "-incomplete"
				h.createRegistrySecret(secretName, registryData.registrySecretData)
				updateData := registryData.toServerlessSpec(secretName)
				shouldUpdateServerless(h, serverlessName, updateData)
				shouldPropagateSpecProperties(h, serverlessRegistrySecret, registryDataIncompleteFilledByDefault)
			}
			{
				registryData := registryDataExternalWithoutSecret
				updateData := registryData.toServerlessSpec("")
				shouldUpdateServerless(h, serverlessName, updateData)
				shouldPropagateSpecProperties(h, serverlessRegistrySecret, registryDataDefault)
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
	Eventually(h.createGetServerlessStateFunc(serverlessName)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 20).
		Should(Equal(v1alpha1.StateReady))
}

func shouldPropagateSpecProperties(h testHelper, registrySecretName string, expected dockerRegistryData) {
	Eventually(h.createCheckRegistrySecretFunc(registrySecretName, expected.registrySecretData)).
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

func (d *dockerRegistryData) toServerlessSpec(secretName string) v1alpha1.ServerlessSpec {
	result := v1alpha1.ServerlessSpec{
		DockerRegistry: &v1alpha1.DockerRegistry{
			EnableInternal: d.EnableInternal,
		},
	}
	if secretName != "" {
		result.DockerRegistry.SecretName = pointer.String(secretName)
	}
	return result
}
