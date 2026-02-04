package function

import (
	"fmt"

	"github.com/kyma-project/serverless/tests/operator/utils"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Create(utils *utils.TestUtils) error {
	function := getFunction(utils)
	if err := utils.Client.Create(utils.Ctx, function); err != nil {
		return fmt.Errorf("failed to create function: %w", err)
	}

	return nil
}

func getFunction(utils *utils.TestUtils) *serverlessv1alpha2.Function {
	return &serverlessv1alpha2.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.FunctionName,
			Namespace: utils.Namespace,
		},
		Spec: serverlessv1alpha2.FunctionSpec{
			Runtime: serverlessv1alpha2.NodeJs24,
			Source: serverlessv1alpha2.Source{
				Inline: &serverlessv1alpha2.InlineSource{
					Source: `module.exports = { 
						main: function(event, context) { 
							return "Hello World"; 
						} 
					}`,
				},
			},
		},
	}
}
