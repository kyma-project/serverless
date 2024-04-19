package serverless

import "github.com/kyma-project/serverless/tests/operator/utils"

func Delete(utils *utils.TestUtils) error {
	serverless := fixServerless(utils)

	return utils.Client.Delete(utils.Ctx, serverless)
}
