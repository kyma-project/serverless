package serverless

import (
	"fmt"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/tests/operator/serverless/deployment"
	"github.com/kyma-project/serverless/tests/operator/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func VerifyDeletion(utils *utils.TestUtils) error {
	return client.IgnoreNotFound(Verify(utils))
}

func Verify(utils *utils.TestUtils) error {
	var serverless v1alpha1.Serverless
	objectKey := client.ObjectKey{
		Name:      utils.ServerlessName,
		Namespace: utils.Namespace,
	}

	if err := utils.Client.Get(utils.Ctx, objectKey, &serverless); err != nil {
		return err
	}

	if err := verifyCondition(utils, &serverless); err != nil {
		return err
	}

	return deployment.VerifyCtrlMngrEnvs(utils, &serverless)
}

func verifyCondition(utils *utils.TestUtils, serverless *v1alpha1.Serverless) error {
	if serverless.Status.State != v1alpha1.StateReady {
		return fmt.Errorf("serverless '%s' in '%s' state", utils.ServerlessName, serverless.Status.State)
	}

	return nil
}
