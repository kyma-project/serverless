package endpoint

import (
	"fmt"
	"net/http"
)

func StartInternalEndpointServer(bindAddr string) {
	http.HandleFunc("/internal/function/", func(w http.ResponseWriter, r *http.Request) {
		// Extract namespaced name from URL
		nsn := r.URL.Path[len("/internal/function/"):]
		// Respond with a dummy payload
		_, _ = fmt.Fprintf(w, "Function lookup for: %s", nsn)
	})

	go func() {
		err := http.ListenAndServe(bindAddr, nil)
		if err != nil {
			panic(fmt.Sprintf("failed to start internal HTTP server: %v", err))
		}
	}()
}
