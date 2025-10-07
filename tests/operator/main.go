package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/kyma-project/serverless/tests/operator/logger"
	"github.com/kyma-project/serverless/tests/operator/namespace"
	"github.com/kyma-project/serverless/tests/operator/serverless"
	"github.com/kyma-project/serverless/tests/operator/utils"
)

var (
	testTimeout = time.Minute * 10
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	log, err := logger.New()
	if err != nil {
		fmt.Printf("%s: %s\n", "unable to setup logger", err)
		os.Exit(1)
	}

	log.Info("Configuring test essentials")
	client, err := utils.GetKuberentesClient()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	log.Info("Start scenario")
	err = runScenario(&utils.TestUtils{
		Namespace: fmt.Sprintf("serverless-test-%s", uuid.New().String()),
		Ctx:       ctx,
		Client:    client,
		Logger:    log,

		ServerlessName:           "default-test",
		ServerlessCtrlDeployName: "serverless-ctrl-mngr",
		ServerlessRegistryName:   "serverless-docker-registry",
		ServerlessUpdateSpec: v1alpha1.ServerlessSpec{
			DockerRegistry: &v1alpha1.DockerRegistry{
				EnableInternal: utils.PtrFromVal(true),
			},
			Tracing: &v1alpha1.Endpoint{
				Endpoint: "http://tracing-endpoint",
			},
			Eventing: &v1alpha1.Endpoint{
				Endpoint: "http://eventing-endpoint",
			},
			TargetCPUUtilizationPercentage:   "10",
			FunctionRequeueDuration:          "19m",
			FunctionBuildExecutorArgs:        "executor-args",
			FunctionBuildMaxSimultaneousJobs: "10",
			HealthzLivenessTimeout:           "20",
			DefaultBuildJobPreset:            "normal",
			DefaultRuntimePodPreset:          "M",
			EnableNetworkPolicies:            true,
		},
	})
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func runScenario(testutil *utils.TestUtils) error {
	// create test namespace
	testutil.Logger.Infof("Creating namespace '%s'", testutil.Namespace)
	if err := namespace.Create(testutil); err != nil {
		return err
	}

	// create Serverless
	testutil.Logger.Infof("Creating legacy serverless '%s'", testutil.ServerlessName)
	if err := serverless.Create(testutil, true); err != nil {
		return err
	}

	// verify Serverless
	testutil.Logger.Infof("Verifying legacy serverless '%s'", testutil.ServerlessName)
	if err := utils.WithRetry(testutil, serverless.Verify); err != nil {
		return err
	}

	// update serverless with other spec
	testutil.Logger.Infof("Updating legacy serverless '%s'", testutil.ServerlessName)
	if err := serverless.Update(testutil); err != nil {
		return err
	}

	// verify Serverless
	testutil.Logger.Infof("Verifying legacy serverless '%s'", testutil.ServerlessName)
	if err := utils.WithRetry(testutil, serverless.Verify); err != nil {
		return err
	}

	// delete Serverless
	testutil.Logger.Infof("Deleting legacy serverless '%s'", testutil.ServerlessName)
	if err := serverless.Delete(testutil, true); err != nil {
		return err
	}

	// verify Serverless deletion
	testutil.Logger.Infof("Verifying legacy serverless '%s' deletion", testutil.ServerlessName)
	if err := utils.WithRetry(testutil, serverless.VerifyDeletion); err != nil {
		return err
	}

	// cleanup
	testutil.Logger.Infof("Deleting namespace '%s'", testutil.Namespace)
	return namespace.Delete(testutil)
}
