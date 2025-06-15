package endpoint

import (
	"net/http"
	"time"

	"encoding/json"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
)

func StartInternalHTTPServer(bindAddr string, log logr.Logger) {
	http.HandleFunc("/internal/function/", handleFunctionRequest)

	go func() {
		err := http.ListenAndServe(bindAddr, nil)
		if err != nil {
			log.Error(err, "failed to start internal HTTP server")
		}
	}()
}

func handleFunctionRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ns := vars["namespace"]
	name := vars["name"]

	resp := map[string]string{
		"namespace": ns,
		"name":      name,
		"time":      time.DateTime,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
