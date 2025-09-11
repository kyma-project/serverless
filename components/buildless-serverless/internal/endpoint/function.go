package endpoint

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/resources"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/endpoint/packagejson"
	"github.com/pkg/errors"
	"go.yaml.in/yaml/v3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) handleFunctionRequest(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")

	s.log.Infof("handling function request for function '%s/%s'", ns, name)

	if ns == "" || name == "" {
		s.writeErrorResponse(w, http.StatusBadRequest, errors.New("missing namespace or name"))
		return
	}

	function := serverlessv1alpha2.Function{}
	err := s.k8s.Get(s.ctx, client.ObjectKey{Namespace: ns, Name: name}, &function)
	if err != nil {
		s.writeErrorResponse(w, http.StatusNotFound, errors.Wrapf(err, "failed to get function '%s/%s'", ns, name))
		return
	}

	svc, err := s.buildServiceFileData(&function)
	if err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, errors.Wrapf(err, "failed to build service for function '%s/%s'", ns, name))
		return
	}

	deployment, err := s.buildDeploymentFileData(&function)
	if err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, errors.Wrapf(err, "failed to build deployment for function '%s/%s'", ns, name))
		return
	}

	runtimeFiles, err := s.getRuntimeFiles(&function)
	if err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, errors.Wrapf(err, "failed to get runtime files for function '%s/%s'", ns, name))
		return
	}

	s.writeFilesListResponse(w, append(runtimeFiles, []FileResponse{
		{Name: "service.yaml", Data: base64.StdEncoding.EncodeToString(svc)},
		{Name: "deployment.yaml", Data: base64.StdEncoding.EncodeToString(deployment)},
	}...))
}

func (s *Server) getRuntimeFiles(function *serverlessv1alpha2.Function) ([]FileResponse, error) {
	if !function.HasNodejsRuntime() {
		// TODO: support non-nodejs runtimes
		return nil, errors.New("ejecting functions with non-nodejs runtimes is not supported")
	}

	runtimeDir := fmt.Sprintf("../runtimes/%s", function.Spec.Runtime)

	packagejsonFile, err := os.ReadFile(runtimeDir + "/package.json")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read package.json")
	}

	packagejsonFile, err = packagejson.Merge([]byte(function.Spec.Source.Inline.Dependencies), packagejsonFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to merge package.json")
	}

	serverFile, err := os.ReadFile(runtimeDir + "/server.mjs")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read server.mjs")
	}

	makefileFile, err := os.ReadFile(runtimeDir + "/cli/Makefile")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read Makefile")
	}

	dockerfileFile, err := os.ReadFile(runtimeDir + "/cli/Dockerfile")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read Dockerfile")
	}

	libFilesInfo, err := os.ReadDir(runtimeDir + "/lib")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read lib directory")
	}

	libFiles := make([]FileResponse, 0, len(libFilesInfo))
	for _, f := range libFilesInfo {
		data, err := os.ReadFile(runtimeDir + "/lib/" + f.Name())
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read lib file '%s'", f.Name())
		}
		libFiles = append(libFiles, FileResponse{Name: fmt.Sprintf("/lib/%s", f.Name()), Data: base64.StdEncoding.EncodeToString(data)})
	}

	return append(libFiles, []FileResponse{
		{Name: "package.json", Data: base64.StdEncoding.EncodeToString(packagejsonFile)},
		{Name: "server.mjs", Data: base64.StdEncoding.EncodeToString(serverFile)},
		{Name: "Dockerfile", Data: base64.StdEncoding.EncodeToString(dockerfileFile)},
		{Name: "handler.js", Data: base64.StdEncoding.EncodeToString([]byte(function.Spec.Source.Inline.Source))},
		{Name: "Makefile", Data: base64.StdEncoding.EncodeToString(makefileFile)},
	}...), nil
}

func (s *Server) buildServiceFileData(function *serverlessv1alpha2.Function) ([]byte, error) {
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

func (s *Server) buildDeploymentFileData(function *serverlessv1alpha2.Function) ([]byte, error) {
	if function.HasGitSources() {
		// TODO: support git source
		return nil, errors.New("ejecting functions with git source is not supported")
	}

	deployName := fmt.Sprintf("%s-ejected", function.Name)
	deploy := resources.NewDeployment(
		function,
		&s.functionConfig, nil,
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
