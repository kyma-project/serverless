package utils

import (
	"context"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TestUtils struct {
	Ctx    context.Context
	Logger *zap.SugaredLogger
	Client client.Client

	Namespace                string
	ServerlessName           string
	SecondServerlessName     string
	FunctionName             string
	ServerlessCtrlDeployName string
	ServerlessConfigName     string
	ServerlessRegistryName   string
	ServerlessConfigMapName  string
	ServerlessUpdateSpec     v1alpha1.ServerlessSpec
}
