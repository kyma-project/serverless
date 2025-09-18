package endpoint

import (
	"net/http"

	"github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/endpoint/runtime"
	"github.com/pkg/errors"
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

	function := v1alpha2.Function{}
	err := s.k8s.Get(s.ctx, client.ObjectKey{Namespace: ns, Name: name}, &function)
	if err != nil {
		s.writeErrorResponse(w, http.StatusNotFound, errors.Wrapf(err, "failed to get function '%s/%s'", ns, name))
		return
	}

	resourceFiles, err := runtime.BuildResources(&s.functionConfig, &function)
	if err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, errors.Wrapf(err, "failed to get resource files for function '%s/%s'", ns, name))
		return
	}

	runtimeFiles, err := runtime.ReadFiles(&function)
	if err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, errors.Wrapf(err, "failed to get runtime files for function '%s/%s'", ns, name))
		return
	}

	s.writeFilesListResponse(w, append(resourceFiles, runtimeFiles...), getOutputMessage())
}

func getOutputMessage() string {
	return "Proposed code structure contains:\n" +
		"- functions code and dependencies\n" +
		"- server code with its build-in functionalities (like cloudevents or tracing)\n" +
		"- resources required to deploy application on the cluster\n" +
		"- scripts and automations to easily manage the application lifecycle\n" +
		"\n" +
		"Read more about next steps and possibilities in the 'README.md' file.\n\n"
}
