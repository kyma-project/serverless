package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kyma-project/serverless/tests/operator/logger"
	"github.com/kyma-project/serverless/tests/operator/utils"
	"github.com/kyma-project/serverless/tests/rbac/clusterrole"
)

var testTimeout = time.Minute * 10

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	log, err := logger.New()
	if err != nil {
		fmt.Printf("unable to setup logger: %s\n", err)
		os.Exit(1)
	}

	// client should be in edit ClusterRole
	client, err := utils.GetKuberentesClient()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	u := &utils.TestUtils{Ctx: ctx, Client: client, Logger: log}

	if err := runScenario(u); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

// runScenario verifies that kyma-serverless-edit and kyma-functions-edit ClusterRoles
// are correctly aggregated into the built-in "edit" ClusterRole and have correct permissions and restrictions.
func runScenario(u *utils.TestUtils) error {
	ns := os.Getenv("RBAC_TEST_NAMESPACE")
	if ns == "" {
		return fmt.Errorf("RBAC_TEST_NAMESPACE env var not set")
	}

	// verify that edit Cluster Role allows to manage functions and serverlesses
	u.Logger.Infof("Verifying edit Cluster Role for functions and serverless in namespace: %s", ns)
	if err := clusterrole.VerifyCRUD(u.Ctx, u.Client, ns, u.Logger); err != nil {
		return err
	}

	// verify that edit Cluster Role forbids creating namespaces (cluster-scoped resource)
	u.Logger.Infof("Verifying cluster-scoped resources limitations for edit Cluster Role")
	if err := clusterrole.VerifyCannotCreateNamespace(u.Ctx, u.Client, u.Logger); err != nil {
		return fmt.Errorf("namespace creation should be forbidden: %w", err)
	}

	u.Logger.Info("All Cluster Role tests passed")
	return nil
}
