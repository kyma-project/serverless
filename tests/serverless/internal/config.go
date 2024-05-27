package internal

import (
	"time"

	"github.com/kyma-project/serverless/tests/serverless/internal/executor"
)

const (
	TestDataKey = "testData"
)

type Config struct {
	Namespace           string               `envconfig:"default=test-function"`
	KubectlProxyEnabled bool                 `envconfig:"default=false"`
	Verbose             bool                 `envconfig:"default=false"`
	WaitTimeout         time.Duration        `envconfig:"default=15m"`
	MaxPollingTime      time.Duration        `envconfig:"default=5m"`
	InsecureSkipVerify  bool                 `envconfig:"default=true"`
	Cleanup             executor.CleanupMode `envconfig:"default=yes"`
	// change gitserver version to main after fixing the issue:
	// https://github.com/kyma-project/serverless/issues/998
	GitServerImage                  string `envconfig:"default=europe-docker.pkg.dev/kyma-project/prod/gitserver:c78bc4083a3c30409784ce6a874165c8a60c00e6"`
	GitServerRepoName               string `envconfig:"default=function"`
	IstioEnabled                    bool   `envconfig:"default=false"`
	PackageRegistryConfigSecretName string `envconfig:"default=serverless-package-registry-config"`
	PackageRegistryConfigURLNode    string `envconfig:"default=https://pkgs.dev.azure.com/kyma-wookiee/public-packages/_packaging/public-packages%40Release/npm/registry/"`
	PackageRegistryConfigURLPython  string `envconfig:"default=https://pkgs.dev.azure.com/kyma-wookiee/public-packages/_packaging/public-packages%40Release/pypi/simple/"`
}
