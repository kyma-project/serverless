package runtime

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/config"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/resources"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/endpoint/types"
	"github.com/pkg/errors"
	"go.yaml.in/yaml/v3"
)

func BuildResources(functionConfig *config.FunctionConfig, f *v1alpha2.Function, appName string) ([]types.FileResponse, error) {
	svc, err := buildServiceFileData(f, appName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to build service")
	}

	deployment, err := buildDeploymentFileData(functionConfig, f, appName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to build deployment")
	}

	return []types.FileResponse{
		{Name: "k8s/service.yaml", Data: base64.StdEncoding.EncodeToString(svc)},
		{Name: "k8s/deployment.yaml", Data: base64.StdEncoding.EncodeToString(deployment)},
	}, nil
}

func buildServiceFileData(function *v1alpha2.Function, appName string) ([]byte, error) {
	svcName := appName
	if svcName == "" {
		svcName = fmt.Sprintf("%s-ejected", function.Name)
	}

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

func buildDeploymentFileData(functionConfig *config.FunctionConfig, function *v1alpha2.Function, appName string) ([]byte, error) {
	if function.HasGitSources() {
		// TODO: support git source
		return nil, errors.New("ejecting functions with git source is not supported")
	}

	deployName := appName
	if deployName == "" {
		deployName = fmt.Sprintf("%s-ejected", function.Name)
	}

	deploy := resources.NewDeployment(
		function,
		functionConfig,
		nil,
		"",
		nil,
		appName,
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

	yamlBytes := bytes.Buffer{}
	e := yaml.NewEncoder(&yamlBytes)
	e.SetIndent(2)
	e.Encode(jsonObj)
	err = e.Encode(jsonObj)
	if err != nil {
		return nil, err
	}

	return yamlBytes.Bytes(), nil
}
