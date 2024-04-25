# Testing Strategy

## CI/CD Jobs Running on Pull Requests

Each pull request to the repository triggers the following CI/CD jobs that verify the Serverless Operator reconciliation logic and run integration tests of the Serverless module:

- `Markdown / documentation-link-check (pull_request)` - Checks if there are no broken links in the pull request `.md` files. For the configuration, see the [mlc.config.json](https://github.com/kyma-project/serverless/blob/main/.mlc.config.json) and the [markdown.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/markdown.yaml) files.
- `Operator verify / operator-lint (pull_request)` - Is responsible for the Operator linting and static code analysis. For the configuration, see the [operator-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/operator-verify.yaml) file.
- `Serverless verify / serverless-lint (pull_request)` - Is responsible for the Serverless linting and static code analysis. For the configuration, see the [serverless-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/serverless-verify.yaml) file.
- `Operator verify / operator-unit-tests (pull_request)` - Runs basic unit tests of Operator's logic. For the configuration, see the [operator-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/operator-verify.yaml) file.
- `Serverless verify / serverless-unit-tests (pull_request)` - Runs unit tests of Serverless's logic. For the configuration, see the [serverless-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/serverless-verify.yaml) file.
- `Operator verify / operator-integration-test (pull_request)` - Runs the create/update/delete Serverless integration tests in k3d cluster. For the configuration, see the [operator-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/operator-verify.yaml) file.
- `Serverless verify / serverless-integration-test (pull_request)` - Runs the basic functionality integration and the `tracing`, `api-gateway`, and `cloud-event` contract compatibility integration test suite for the Serverless in a k3d cluster. For the configuration, see the [serverless-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/serverless-verify.yaml) file.
- `Gitleaks / gitleaks-scan (pull_request)` - Scans the pull request for secrets and credentials.

## CI/CD Jobs Running on the Main Branch

- `Operator verify / operator-integration-test (push)` - Runs the create/update/delete Serverless integration tests in k3d cluster. For the configuration, see the [operator-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/operator-verify.yaml) file.
- `Serverless verify / serverless-integration-test (push)` - Runs the basic functionality integration and the `tracing`, `api-gateway`, and `cloud-event` contract compatibility integration test suite for the Serverless in a k3d cluster. For the configuration, see the [serverless-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/serverless-verify.yaml) file.
- `Operator verify / operator-upgrade-test (push)` - Runs the upgrade integration test suite and verifies if the latest release can be successfully upgraded to the new (`main`) revision. For the configuration, see the [operator-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/operator-verify.yaml) file.
- `Serverless verify / serverless-upgrade-test (push)` - Runs the basic functionality integration and the `tracing`, `api-gateway`, and `cloud-event` contract compatibility integration test suite for the Serverless in a k3d cluster after upgrading from the latest release to the actual revision to check if the serverless component is working properly after the upgrade. For the configuration, see the [serverless-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/serverless-verify.yaml) file.
- `Serverless verify / git-auth-integration-test (push)` - Runs the `GitHub` and `Azure DevOps` API and authentication integration test suite for Serverless. For the configuration, see the [serverless-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/serverless-verify.yaml) file.
- `Operator verify / gardener-integration-test (push)` - Checks the installation of the Serverless module in the Gardener shoot cluster and runs basic integration tests of Serverless. For the configuration, see the [operator-verify.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/operator-verify.yaml) file.

## CI/CD Jobs Running on a Schedule

- `Markdown / documentation-link-check` - Runs Markdown link check every day at 05:00 AM.
