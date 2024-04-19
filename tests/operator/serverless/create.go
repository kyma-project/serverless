package serverless

import (
	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/tests/operator/utils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Create(utils *utils.TestUtils) error {
	serverlessObj := fixServerless(utils)

	return utils.Client.Create(utils.Ctx, serverlessObj)
}

func fixServerless(testUtils *utils.TestUtils) *v1alpha1.Serverless {
	return &v1alpha1.Serverless{
		ObjectMeta: v1.ObjectMeta{
			Name:      testUtils.ServerlessName,
			Namespace: testUtils.Namespace,
		},
		Spec: v1alpha1.ServerlessSpec{
			DockerRegistry: &v1alpha1.DockerRegistry{
				EnableInternal: utils.PtrFromVal(false),
			},
		},
	}
}
