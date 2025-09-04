package endpoint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func (s *Server) writeErrorResponse(w http.ResponseWriter, status int, respErr error) {
	headerStatus := status

	buf := bytes.NewBuffer([]byte{})
	err := json.NewEncoder(buf).Encode(ErrorResponse{Error: respErr.Error()})
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

type FileResponse struct {
	Name string `json:"name"`
	Data string `json:"data"` // base64 encoded file content
}

type FilesListResponse struct {
	Files []FileResponse `json:"files"`
}

func (s *Server) writeFilesListResponse(w http.ResponseWriter, data []FileResponse) {
	buf := bytes.NewBuffer([]byte{})
	err := json.NewEncoder(buf).Encode(FilesListResponse{Files: data})
	if err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, errors.Wrap(err, "failed to encode response"))
		return
	}

	s.log.Debugf("writing item list response with %d items", len(data))
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, buf.String())
}
