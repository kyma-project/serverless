# Testing Strategy

Each pull request to the repository triggers CI/CD jobs that verify the Serverless Operator reconciliation logic and run integration tests of the Serverless module.

- `pre-serverless-operator-operator-build` - Compiles Serverless Operator's code and pushes its Docker image.
- `pre-serverless-operator-operator-tests` - Tests the Serverless Operator reconciliation code (Serverless CR CRUD operations).
- `pre-main-serverless-operator-verify` - Performs integration testing for the Serverless module installed by Serverless Operator (not using Lifecycle Manager). This job includes [contract tests](https://github.com/kyma-project/kyma/issues/17501) to confirm that contracts towards other Kyma modules or industry standard specifications are not broken.
- `pull-serverless-module-build` - Bundles a ModuleTemplate manifest that allows testing it against Lifecycle Manager manually. 

After the pull request is merged, the following CI/CD jobs are executed:

 - `post-main-serverless-operator-verify` - Installs the Serverless module (using Lifecycle Manager) and runs integration tests of Serverless.
 - `post-serverless-operator-build` - Rebuilds the Serverless module and the ModuleTemplate manifest file that can be submitted to modular Kyma.
 - `post-main-serverless-operator-upgrade-latest-to-main` - Installs the latest released version of the Serverless module, upgrades to the version from main, and runs integration tests.
 - `gardener-integration-test` - Installs the Serverless module (not using Lifecycle Manager) on the Gardener shoot cluster and runs integration tests of Serverless.
