package namespace

import "github.com/kyma-project/serverless/tests/operator/utils"

func Delete(utils *utils.TestUtils) error {
	namespace := fixNamespace(utils)

	return utils.Client.Delete(utils.Ctx, namespace)
}
