package endpoint

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func InternalDummyHandler(w http.ResponseWriter, r *http.Request) {
	log := zap.SugaredLogger{}
	vars := mux.Vars(r)
	ns := vars["namespace"]
	name := vars["name"]

	resp := map[string]string{
		"namespace": ns,
		"name":      name,
		"time":      time.DateTime,
	}

	log.Info(fmt.Sprintf("------------test-log--------------"))
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Error(fmt.Sprintf("failed to encode response: %v", err))
	}
}
