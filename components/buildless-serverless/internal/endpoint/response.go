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
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")

	buf := bytes.NewBuffer([]byte{})
	err := json.NewEncoder(buf).Encode(ErrorResponse{Error: respErr.Error()})
	if err != nil {
		// If encoding fails, we can't do much, so we log the error and return a generic error response
		s.log.Errorf("failed to encode error response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		buf = bytes.NewBufferString(`{"error":"internal server error"}`)
	}

	fmt.Fprint(w, buf.String())
}

type ItemListResponse struct {
	Items []interface{} `json:"items"`
}

func (s *Server) writeItemListResponse(w http.ResponseWriter, data []interface{}) {
	buf := bytes.NewBuffer([]byte{})
	err := json.NewEncoder(buf).Encode(ItemListResponse{Items: data})
	if err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, errors.Wrap(err, "failed to encode response"))
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, buf.String())
}
