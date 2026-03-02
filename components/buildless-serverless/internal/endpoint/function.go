package endpoint

import (
	"net/http"
	"strings"

	"github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/endpoint/runtime"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) handleFunctionRequest(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	appName := r.URL.Query().Get("targetAppName")

	s.log.Infof("handling function request for function '%s/%s'", ns, name)

	if err := validateFunctionParams(ns, name, appName); err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	function := v1alpha2.Function{}
	err := s.k8s.Get(s.ctx, client.ObjectKey{Namespace: ns, Name: name}, &function)
	if err != nil {
		s.writeErrorResponse(w, http.StatusNotFound, errors.Wrapf(err, "failed to get function '%s/%s'", ns, name))
		return
	}

	resourceFiles, err := runtime.BuildResources(&s.functionConfig, &function, appName, s.isKymaFipsModeEnabled)
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

func validateFunctionParams(ns string, name string, appName string) error {
	if ns == "" || name == "" {
		return errors.New("missing namespace or name")
	}
	for paramName, paramValue := range map[string]string{
		"namespace":     ns,
		"name":          name,
		"targetAppName": appName,
	} {
		if paramValue == "" {
			continue
		}
		if errs := validation.IsDNS1123Label(paramValue); len(errs) > 0 {
			return errors.Wrapf(errors.New(strings.Join(errs, "; ")),
				"invalid parameter %q", paramName)
		}
	}
	return nil
}

func getOutputMessage() string {
	return "The proposed code structure contains:\n" +
		"- functions code and dependencies\n" +
		"- server code with its built-in functionalities (like cloudevents or tracing)\n" +
		"- resources required to deploy the application on the cluster\n" +
		"- scripts and automations to easily manage the application lifecycle\n" +
		"\n" +
		"Read more about next steps and possibilities in the 'README.md' file.\n\n"
}
