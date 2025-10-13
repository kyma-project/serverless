package serverless

import (
	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/tests/operator/utils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Create(utils *utils.TestUtils) error {
	serverlessObj := fixServerless(utils, utils.ServerlessName)

	return utils.Client.Create(utils.Ctx, serverlessObj)
}

func CreateSecond(utils *utils.TestUtils) error {
	serverlessObj := fixServerless(utils, utils.SecondServerlessName)

	return utils.Client.Create(utils.Ctx, serverlessObj)
}

func fixServerless(testUtils *utils.TestUtils, name string) *v1alpha1.Serverless {
	annotations := map[string]string{}
	if testUtils.LegacyMode {
		annotations["serverless.kyma-project.io/buildless-mode"] = "disabled"
	}

	return &v1alpha1.Serverless{
		ObjectMeta: v1.ObjectMeta{
			Name:        name,
			Namespace:   testUtils.Namespace,
			Annotations: annotations,
		},
		Spec: v1alpha1.ServerlessSpec{
			DockerRegistry: &v1alpha1.DockerRegistry{
				EnableInternal: utils.PtrFromVal(false),
			},
		},
	}
}
