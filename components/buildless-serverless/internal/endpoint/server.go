package endpoint

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Server struct {
	ctx context.Context
	mux *mux.Router
	k8s client.Client
	log *zap.SugaredLogger
}

func NewInternalServer(ctx context.Context, k8s client.Client, log *zap.SugaredLogger) *Server {
	server := &Server{
		ctx: ctx,
		mux: mux.NewRouter(),
		k8s: k8s,
		log: log,
	}

	server.mux.HandleFunc("/internal/function/", server.handleFunctionRequest)

	return server
}

func (s *Server) ListenAndServe(bindAddr string) error {
	return http.ListenAndServe(bindAddr, s.mux)
}
