# Testing Strategy

Each pull request to the repository triggers CI/CD jobs that verify the Serverless Operator reconciliation logic and run integration tests of the Serverless module.

- `pre-serverless-operator-operator-build` - Compiling the Serverless Operator code and pushing its docker image.
- `pre-serverless-operator-operator-tests` - Testing the Serverless Operator reconciliation code (Serverless CR CRUD operations).
- `pre-main-serverless-operator-verify` - Integration testing for the Serverless module installed by Serverless Operator (not using Lifecycle Manager).
- `pull-serverless-module-build` - Bundling a module template manifest that allows testing it against Lifecycle Manager manually. 

After the pull request is merged, the following CI/CD jobs are executed:

 - `post-main-serverless-operator-verify` - Installs the Serverless module (using Lifecycle Manager) and runs integration tests of Serverless.
 - `post-serverless-operator-build` - rebuilds the Serverless module and the module template manifest file that can be submitted to modular Kyma.