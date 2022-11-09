package controllers

import (
	"context"
	"fmt"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

var _ = Describe("Serverless controller", func() {
	Context("When creating fresh instance", func() {
		const (
			namespaceName  = "kyma-system"
			serverlessName = "test"
			operatorName   = "keda-manager"
		)

		var (
			serverlessSpec = v1alpha1.ServerlessSpec{
				DockerRegistry: &v1alpha1.DockerRegistry{
					EnableInternal: pointer.Bool(true),
				},
			}
		)

		It("The status should be Success", func() {
			h := testHelper{
				ctx:           context.Background(),
				namespaceName: namespaceName,
			}
			h.createNamespace()

			// operations like C(R)UD can be tested in separated tests,
			// but we have time-consuming flow and decided do it in one test
			shouldCreateServerless(h, serverlessName, serverlessSpec)

			// TODO: finish use-case
			// shouldPropagateKedaCrdSpecProperties(h, kedaDeploymentName, metricsDeploymentName, kedaSpec)

			// TODO: disabled because of bug in operator (https://github.com/kyma-project/module-manager/issues/94)
			// shouldUpdateKeda(h, kedaName, kedaDeploymentName)

			// shouldDeleteKeda(h, kedaName)
		})
	})
})

func shouldCreateServerless(h testHelper, serverlessName string, spec v1alpha1.ServerlessSpec) {
	// act
	h.createServerless(serverlessName, spec)

	// TODO: we have to update deployment status manually
	// h.updateDeploymentStatus(metricsDeploymentName)
	// h.updateDeploymentStatus(kedaDeploymentName)

	// TODO: assert
	// Eventually(h.createGetKedaStateFunc(kedaName)).
	// 	WithPolling(time.Second * 2).
	// 	WithTimeout(time.Second * 20).
	// 	Should(Equal(rtypes.StateReady))
}

type testHelper struct {
	ctx           context.Context
	namespaceName string
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
