# Testing Strategy

Each pull request to the repository triggers CI/CD jobs that verify the Serverless Operator reconciliation logic and run integration tests of the Serverless module.

- `pre-serverless-operator-operator-build` - Compiles Serverless Operator's code and pushes its Docker image.
- `pre-serverless-operator-operator-tests` - Tests the Serverless Operator reconciliation code (Serverless CR CRUD operations).
- `pre-main-serverless-operator-verify` - Performs integration testing for the Serverless module installed by Serverless Operator (not using Lifecycle Manager).
- `pull-serverless-module-build` - Bundles a ModuleTemplate manifest that allows testing it against Lifecycle Manager manually. 

After the pull request is merged, the following CI/CD jobs are executed:

 - `post-main-serverless-operator-verify` - Installs the Serverless module (using Lifecycle Manager) and runs integration tests of Serverless.
 - `post-serverless-operator-build` - Rebuilds the Serverless module and the ModuleTemplate manifest file that can be submitted to modular Kyma.