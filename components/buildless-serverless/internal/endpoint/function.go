package endpoint

import (
	"net/http"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/git"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/resources"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) handleFunctionRequest(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
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

	svc := resources.NewService(&function, serverlessv1alpha2.TrimClusterInfoLabels)

	deployment, err := s.buildDeployment(&function)
	if err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, errors.Wrapf(err, "failed to build deployment for function '%s/%s'", ns, name))
		return
	}

	s.writeItemListResponse(w, []interface{}{svc, deployment})
}

func (s *Server) buildDeployment(function *serverlessv1alpha2.Function) (*resources.Deployment, error) {
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

	deploy := resources.NewDeployment(function, &s.functionConfig, nil, commit, gitAuth, serverlessv1alpha2.TrimClusterInfoLabels)

	// set strict name
	deploy.SetName(function.GetName())
	deploy.GenerateName = ""

	return deploy, nil
}
