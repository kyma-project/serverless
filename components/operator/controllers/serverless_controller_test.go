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
)

var _ = Describe("Serverless controller", func() {
	Context("When creating fresh instance", func() {
		const (
			namespaceName = "kyma-system"
		)

		var (
			serverlessDataDefault = serverlessData{
				TraceCollectorURL: ptr.To[string](v1alpha1.EndpointDisabled),
			}
			serverlessDataWithChangedDependencies = serverlessData{
				EventPublisherProxyURL: ptr.To[string]("test-eventing-address"),
				TraceCollectorURL:      ptr.To[string]("test-tracing-address"),
			}
		)

		It("The status should be Success", func() {
			h := testHelper{
				ctx:           context.Background(),
				namespaceName: namespaceName,
			}
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
