package serverless

import "github.com/kyma-project/serverless/tests/operator/utils"

func Delete(utils *utils.TestUtils, legacy bool) error {
	serverless := fixServerless(utils, legacy)

	return utils.Client.Delete(utils.Ctx, serverless)
}
