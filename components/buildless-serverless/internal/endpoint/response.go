package endpoint

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

func (s *Server) writeErrorResponse(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"error": "%s"}`, err.Error())
}

func (s *Server) writeItemListResponse(w http.ResponseWriter, data []interface{}) {
	resp := functionResponse{
		Items: data,
	}

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, errors.Wrap(err, "failed to encode response"))
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
}
