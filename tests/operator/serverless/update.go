package serverless

import (
	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/tests/operator/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Update(testutils *utils.TestUtils) error {
	var serverless v1alpha1.Serverless
	objectKey := client.ObjectKey{
		Name:      testutils.SecondServerlessName,
		Namespace: testutils.Namespace,
	}

	if err := testutils.Client.Get(testutils.Ctx, objectKey, &serverless); err != nil {
		return err
	}

	serverless.Spec = testutils.ServerlessUpdateSpec

	return testutils.Client.Update(testutils.Ctx, &serverless)
}
