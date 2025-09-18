package runtime

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/config"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/resources"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/endpoint/types"
	"github.com/pkg/errors"
	"go.yaml.in/yaml/v2"
)

func BuildResources(functionConfig *config.FunctionConfig, f *v1alpha2.Function) ([]types.FileResponse, error) {
	svc, err := buildServiceFileData(f)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to build service")
	}

	deployment, err := buildDeploymentFileData(functionConfig, f)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to build deployment")
	}

	return []types.FileResponse{
		{Name: "resources/service.yaml", Data: base64.StdEncoding.EncodeToString(svc)},
		{Name: "resources/deployment.yaml", Data: base64.StdEncoding.EncodeToString(deployment)},
	}, nil
}

func buildServiceFileData(function *v1alpha2.Function) ([]byte, error) {
	svcName := fmt.Sprintf("%s-ejected", function.Name)
	svc := resources.NewService(
		function,
		resources.ServiceName(svcName),
		resources.ServiceTrimClusterInfoLabels(),
		resources.ServiceAppendSelectorLabels(map[string]string{
			"app.kubernetes.io/instance": svcName,
		}),
	).Service

	data, err := convertK8SObjectToYaml(svc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal service to YAML")
	}

	return data, nil
}

func buildDeploymentFileData(functionConfig *config.FunctionConfig, function *v1alpha2.Function) ([]byte, error) {
	if function.HasGitSources() {
		// TODO: support git source
		return nil, errors.New("ejecting functions with git source is not supported")
	}

	deployName := fmt.Sprintf("%s-ejected", function.Name)
	deploy := resources.NewDeployment(
		function,
		functionConfig,
		nil,
		"",
		nil,
		resources.DeploySetName(deployName),
		resources.DeployTrimClusterInfoLabels(),
		resources.DeployAppendSelectorLabels(map[string]string{
			"app.kubernetes.io/instance": deployName,
		}),
		resources.DeploySetCmd([]string{}), // clear the command to use the default one from the image
		resources.DeploySetImage("image:tag"),
		resources.DeployUseGeneralEnvs(),
	).Deployment

	data, err := convertK8SObjectToYaml(deploy)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal deployment to YAML")
	}

	return data, nil
}

// k8s object are designed to be converted to JSON instead of YAML
// this function does double convertion (from obj to json and from json to yaml)
func convertK8SObjectToYaml(obj interface{}) ([]byte, error) {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var jsonObj interface{}

	// We are using yaml.Unmarshal here (instead of json.Unmarshal) because the
	// Go JSON library doesn't try to pick the right number type (int, float, etc.)
	err = yaml.Unmarshal(jsonBytes, &jsonObj)
	if err != nil {
		return nil, err
	}

	yamlBytes, err := yaml.Marshal(jsonObj)
	if err != nil {
		return nil, err
	}

	return yamlBytes, nil
}
