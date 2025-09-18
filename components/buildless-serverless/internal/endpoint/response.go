package endpoint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kyma-project/serverless/components/buildless-serverless/internal/endpoint/types"
	"github.com/pkg/errors"
)

func (s *Server) writeErrorResponse(w http.ResponseWriter, status int, respErr error) {
	headerStatus := status

	buf := bytes.NewBuffer([]byte{})
	err := json.NewEncoder(buf).Encode(types.ErrorResponse{Error: respErr.Error()})
	if err != nil {
		// If encoding fails, we can't do much, so we log the error and return a generic error response
		s.log.Errorf("failed to encode error response: %v", err)
		headerStatus = http.StatusInternalServerError
		buf = bytes.NewBufferString(`{"error":"internal server error"}`)
	}

	s.log.Debugf("writing error response with status: %d", headerStatus)
	w.WriteHeader(headerStatus)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, buf.String())
}

func (s *Server) writeFilesListResponse(w http.ResponseWriter, data []types.FileResponse) {
	buf := bytes.NewBuffer([]byte{})
	err := json.NewEncoder(buf).Encode(types.FilesListResponse{Files: data})
	if err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, errors.Wrap(err, "failed to encode response"))
		return
	}

	s.log.Debugf("writing item list response with %d items", len(data))
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, buf.String())
}
