package function

import "github.com/kyma-project/serverless/tests/operator/utils"

func Delete(utils *utils.TestUtils) error {
	function := getFunction(utils)
	return utils.Client.Delete(utils.Ctx, function)
}
