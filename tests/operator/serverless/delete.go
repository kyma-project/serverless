package serverless

import "github.com/kyma-project/serverless/tests/operator/utils"

func DeleteOld(utils *utils.TestUtils) error {
	serverless := fixServerless(utils, utils.ServerlessName)

	return utils.Client.Delete(utils.Ctx, serverless)
}
func DeleteNew(utils *utils.TestUtils) error {
	serverless := fixServerless(utils, utils.SecondServerlessName)

	return utils.Client.Delete(utils.Ctx, serverless)
}
