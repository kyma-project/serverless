# Testing Strategy

Each pull request to the repository triggers the following CI/CD jobs that verify the Serverless Operator reconciliation logic and run integration tests of the Serverless module:

- `Markdown / link-check (pull_request)` - Checks if there are no broken links in the pull request `.md` files. For the configuration, see the [mlc.config.json](https://github.com/kyma-project/serverless/blob/main/.mlc.config.json) file.
- `Operator verify / lint (pull_request)` - Is responsible for the Operator linting and static code analysis. For the configuration, see the [operator-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/operator-verify.yaml#L19) file.
- `Serverless verify / lint (pull_request)` - Is responsible for the Serverless linting and static code analysis. For the configuration, see the [serverless-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/serverless-verify.yaml#L19) file.
- `Operator verify / unit-test (pull_request)` - Executes basic create/update/delete functional tests of Operator's reconciliation logic. For the configuration, see the [operator-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/operator-verify.yaml#L30) file.
- `Serverless verify / unit-test (pull_request)` - Executes basic create/update/delete functional tests of Serverless's reconciliation logic. For the configuration, see the [serverless-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/serverless-verify.yaml#L31) file.
- `Operator verify / integration-test (pull_request)` - Executes the main integration test suite for the Operator on a k3d cluster. For the configuration, see the [operator-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/operator-verify.yaml#L39) file.
- `Serverless verify / contract-integration-test (pull_request)` - Executes the contract integration test suite for the Serverless on a k3d cluster. For the configuration, see the [serverless-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/serverless-verify.yaml#L41) file.
- `Operator verify / upgrade-test (pull_request)` - Executes the upgrade integration test suite and verifies if the existing release can be successfully upgraded to the new version. For the configuration, see the [operator-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/operator-verify.yaml#L66) file.
- `Serverless verify / git-auth-integration-test (pull_request)` - Executes the Git authentication test suite for Serverless. For the configuration, see the [serverless-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/serverless-verify.yaml#L64) file.
- `Operator verify / gardener-integration-test (pull_request)` - Installs the Serverless module (not using Lifecycle Manager) on the Gardener shoot cluster and runs integration tests of Serverless. For the configuration, see the [operator-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/operator-verify.yaml#L86) file.
- `pre-serverless-controller-build` - Compiles Serverless Controller's code and pushes its Docker image.
- `pre-serverless-jobinit-build` - Checks the configuration needed for initializing Serverless.
- `pre-serverless-operator-build` - Compiles Serverless Operator's code and pushes its Docker image.
- `pre-serverless-runtimes-java17-jvm-alpha-build` - Runs tests specific to Java 17.
- `pre-serverless-runtimes-nodejs-v16-build` - Runs tests specific to Node.js 16.
- `pre-serverless-runtimes-nodejs-v18-build` - Runs tests specific to Node.js 18.
- `pre-serverless-runtimes-python39-build` - Runs tests specific to Python 3.9.
- `pre-serverless-webhook-build` - Checks the compatibility with the Serverless architecture and ensures that the webhook functionality is working as expected.

After the pull request is merged, the following CI/CD jobs are executed:

 - `Operator verify / lint (push)` - Is responsible for the Operator linting and static code analysis. For the configuration, see the [operator-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/operator-verify.yaml#L19) file.
- `Serverless verify / lint (push)` - Is responsible for the Serverless linting and static code analysis. For the configuration, see the [serverless-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/serverless-verify.yaml#L19) file.
- `Operator verify / unit-test (push)` - Executes basic create/update/delete functional tests of Operator's reconciliation logic. For the configuration, see the [operator-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/operator-verify.yaml#L30) file.
- `Serverless verify / unit-test (push)` - Executes basic create/update/delete functional tests of Serverless's reconciliation logic. For the configuration, see the [serverless-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/serverless-verify.yaml#L31) file.
- `Operator verify / integration-test (push)` - Executes the main integration test suite for the Operator on a k3d cluster. For the configuration, see the [operator-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/operator-verify.yaml#L39) file.
- `Serverless verify / contract-integration-test (push)` - Executes the contract integration test suite for the Serverless on a k3d cluster. For the configuration, see the [serverless-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/serverless-verify.yaml#L41) file.
- `Operator verify / upgrade-test (push)` - Executes the upgrade integration test suite and verifies if the existing release can be successfully upgraded to the new version. For the configuration, see the [operator-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/operator-verify.yaml#L66) file.
- `Serverless verify / git-auth-integration-test (push)` - Executes the Git authentication test suite for Serverless. For the configuration, see the [serverless-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/serverless-verify.yaml#L64) file.
- `Operator verify / gardener-integration-test (push)` - Checks the innstalltion of the Serverless module (not using Lifecycle Manager) on the Gardener shoot cluster and runs integration tests of Serverless. For the configuration, see the [operator-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/operator-verify.yaml#L86) file.
- `post-serverless-controller-build` - Re-builds the Controller's image and pushes it into the ``prod` registry.
- `post-serverless-jobinit-build` - Verifies the completion of the Serverless build.
- `post-serverless-operator-build` - Rebuilds the Serverless module and the ModuleTemplate manifest file that can be submitted to modular Kyma.
- `post-serverless-runtimes-java17-jvm-alpha-build` - Runs tests specific to Java 17 after the build.
- `post-serverless-runtimes-nodejs-v16-build` - Runs tests specific to Node.js 16 after the build.
- `post-serverless-runtimes-nodejs-v18-build` - Runs tests specific to Node.js 18 after the build.
- `post-serverless-runtimes-python39-build` - Runs tests specific to Python 3.9 after the build.
- `post-serverless-webhook-build` - Verifies the deployment of the webhook.

 - `gardener-integration-test` - Installs the Serverless module (not using Lifecycle Manager) on the Gardener shoot cluster and runs integration tests of Serverless.
