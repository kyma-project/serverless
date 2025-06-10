package endpoint

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

func InternalDummyHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ns := vars["namespace"]
	name := vars["name"]

	resp := map[string]string{
		"namespace": ns,
		"name":      name,
		"time":      time.DateTime,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
