package endpoint

import (
	"bytes"
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
	buf := bytes.NewBuffer([]byte{})
	err := json.NewEncoder(buf).Encode(data)
	if err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, errors.Wrap(err, "failed to encode response"))
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"items": %s}`, buf.String())
}
