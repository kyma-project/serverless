package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kyma-project/serverless/tests/operator/logger"
	"github.com/kyma-project/serverless/tests/operator/utils"
	"github.com/kyma-project/serverless/tests/rbac/clusterrole"
	"github.com/kyma-project/serverless/tests/rbac/resource"
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

// runScenario dispatches to the correct test based on RBAC_TEST_SCENARIO ("edit" or "view").
// Resources are loaded from the YAML file specified by the RESOURCES_FILE env var.
func runScenario(u *utils.TestUtils) error {
	ns := os.Getenv("RBAC_TEST_NAMESPACE")
	if ns == "" {
		return fmt.Errorf("RBAC_TEST_NAMESPACE env var not set")
	}

	scenario := strings.TrimSpace(strings.ToLower(os.Getenv("RBAC_TEST_SCENARIO")))
	if scenario == "" {
		return fmt.Errorf("RBAC_TEST_SCENARIO env var not set")
	}

	resources, err := resolveResources(u)
	if err != nil {
		return err
	}

	switch scenario {
	case "edit":
		return runEditScenario(u, ns, resources)
	case "view":
		return runViewScenario(u, ns, resources)
	case "precreate":
		return runPrecreateScenario(u, ns, resources)
	default:
		return fmt.Errorf("unknown RBAC_TEST_SCENARIO %q (expected 'edit', 'view', or 'precreate')", scenario)
	}
}

// resolveResources loads resources from the YAML file specified by RESOURCES_FILE.
func resolveResources(u *utils.TestUtils) ([]resource.TestResource, error) {
	path := os.Getenv("RESOURCES_FILE")
	if path == "" {
		return nil, fmt.Errorf("RESOURCES_FILE env var not set")
	}
	if !filepath.IsAbs(path) {
		if wd, err := os.Getwd(); err == nil {
			path = filepath.Join(wd, path)
		}
	}

	u.Logger.Infof("loading test resources from %s", path)
	all, err := resource.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("loading resources: %w", err)
	}
	u.Logger.Infof("loaded %d resource(s) from YAML", len(all))
	return all, nil
}

// runEditScenario verifies that the "edit" ClusterRole grants full CRUD on the resources and forbids creating cluster-scoped resources (namespaces).
func runEditScenario(u *utils.TestUtils, ns string, resources []resource.TestResource) error {
	u.Logger.Infof("=== edit scenario: verifying CRUD in namespace %s ===", ns)
	if err := clusterrole.VerifyEditCRUD(u.Ctx, u.Client, ns, resources, u.Logger); err != nil {
		return err
	}

	u.Logger.Info("=== edit scenario: verifying cluster-scoped restrictions ===")
	if err := clusterrole.VerifyCannotCreateNamespace(u.Ctx, u.Client, u.Logger); err != nil {
		return fmt.Errorf("namespace creation should be forbidden: %w", err)
	}

	u.Logger.Info("All edit ClusterRole tests passed ✓")
	return nil
}

// runViewScenario verifies that the "view" ClusterRole allows read (get/list) but forbids create, update, and delete on the resources.
// NOTE: the resources must already exist in the namespace. Use PreCreateResources in a higher-privilege step before switching to the view role and running this scenario.
func runViewScenario(u *utils.TestUtils, ns string, resources []resource.TestResource) error {
	u.Logger.Infof("=== view scenario: verifying read-only access in namespace %s ===", ns)
	if err := clusterrole.VerifyViewReadOnly(u.Ctx, u.Client, ns, resources, u.Logger); err != nil {
		return err
	}

	u.Logger.Info("=== view scenario: verifying cluster-scoped restrictions ===")
	if err := clusterrole.VerifyCannotCreateNamespace(u.Ctx, u.Client, u.Logger); err != nil {
		return fmt.Errorf("namespace creation should be forbidden: %w", err)
	}

	u.Logger.Info("All view ClusterRole tests passed ✓")
	return nil
}

// runPrecreateScenario creates the test resources with so they exist when the view scenario is executed with a lower-privilege kubeconfig.
func runPrecreateScenario(u *utils.TestUtils, ns string, resources []resource.TestResource) error {
	u.Logger.Infof("=== precreate scenario: creating resources in namespace %s ===", ns)
	if err := clusterrole.PreCreateResources(u.Ctx, u.Client, ns, resources, u.Logger); err != nil {
		return err
	}
	u.Logger.Info("All resources pre-created ✓")
	return nil
}
