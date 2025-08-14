package endpoint

import (
	"net/http"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type functionResponse struct {
	Items []interface{} `json:"items"`
}

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

	s.writeItemListResponse(w, []interface{}{function})
}
