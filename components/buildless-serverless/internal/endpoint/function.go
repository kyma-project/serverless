package endpoint

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/git"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/resources"
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

	s.writeFilesListResponse(w, []FileResponse{
		{Name: "service.yaml", Data: base64.StdEncoding.EncodeToString(svc)},
		{Name: "deployment.yaml", Data: base64.StdEncoding.EncodeToString(deployment)},
	})
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
	var commit string
	var gitAuth *git.GitAuth
	if function.Spec.Source.GitRepository != nil {
		var err error
		if function.HasGitAuth() {
			gitAuth, err = git.NewGitAuth(s.ctx, s.k8s, function)
			if err != nil {
				return nil, errors.Wrap(err, "failed to create git auth")
			}
		}

		gitRepo := function.Spec.Source.GitRepository
		commit, err = git.GetLatestCommit(gitRepo.URL, gitRepo.Reference, gitAuth)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get latest commit")
		}
	}

	deployName := fmt.Sprintf("%s-ejected", function.Name)
	deploy := resources.NewDeployment(
		function,
		&s.functionConfig, nil,
		commit,
		gitAuth,
		resources.DeployName(deployName),
		resources.DeployTrimClusterInfoLabels(),
		resources.DeployAppendSelectorLabels(map[string]string{
			"app.kubernetes.io/instance": deployName,
		}),
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
