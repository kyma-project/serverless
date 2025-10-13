package endpoint

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/config"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Server struct {
	ctx            context.Context
	mux            *mux.Router
	k8s            client.Client
	log            *zap.SugaredLogger
	functionConfig config.FunctionConfig
}

func NewInternalServer(ctx context.Context, log *zap.SugaredLogger, k8s client.Client, functionConfig config.FunctionConfig) *Server {
	server := &Server{
		ctx:            ctx,
		mux:            mux.NewRouter(),
		k8s:            k8s,
		log:            log,
		functionConfig: functionConfig,
	}

	server.mux.HandleFunc("/internal/function/eject/", server.handleFunctionRequest)

	return server
}

func (s *Server) ListenAndServe(bindAddr string) error {
	return http.ListenAndServe(bindAddr, s.mux)
}
