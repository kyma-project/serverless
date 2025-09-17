# Testing Strategy

## CI/CD Jobs Running on Pull Requests

Each pull request to the repository triggers the following CI/CD jobs that verify the Serverless Operator reconciliation logic and run integration tests of the Serverless module:

- `lint / operator-lint` - Is responsible for the Operator linting and static code analysis. For the configuration, see the [lint.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/lint.yaml) file.
- `lint / serverless-lint` - Is responsible for the Serverless linting and static code analysis. For the configuration, see the [lint.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/lint.yaml) file.
- `pull / unit tests / operator-unit-tests` - Runs basic unit tests of Operator's logic. For the configuration, see the [_unit-tests.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/_unit-tests.yaml) file.
- `pull / unit tests / serverless-unit-tests` - Runs unit tests of Serverless's logic. For the configuration, see the [_unit-tests.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/_unit-tests.yaml) file.
- `pull / integrations / operator-integration-test` - Runs the create/update/delete Serverless integration tests in k3d cluster. For the configuration, see the [_integration-tests.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/_integration-tests.yaml) file.
- `pull / integrations / serverless-integration-test` - Runs the basic functionality integration and the `tracing`, `api-gateway`, and `cloud-event` contract compatibility integration test suite for the Serverless in a k3d cluster. For the configuration, see the [_integration-tests.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/_integration-tests.yaml) file.
- `pull / gitleaks / gitleaks-scan` - Scans the pull request for secrets and credentials. For the configuration, see the [_gitleaks.yaml.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/_gitleaks.yaml) file.

## CI/CD Jobs Running on the Main Branch

- `markdown / documentation-link-check` - Checks if there are no broken links in `.md` files. For the configuration, see the [mlc.config.json](https://github.com/kyma-project/serverless/blob/main/.mlc.config.json) and the [markdown.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/markdown.yaml) files.
- `push / integrations / serverless-integration-test` - Runs the create/update/delete Serverless integration tests in k3d cluster. For the configuration, see the [_integration-tests.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/_integration-tests.yaml) file.
- `integration tests (push) / serverless-integration-test` - Runs the basic functionality integration and the `tracing`, `api-gateway`, `cloud-event` and `hana-client` contract compatibility integration test suite for Serverless in a k3d cluster. For the configuration, see the [_integration-tests.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/_integration-tests.yaml) file.
- `push / integrations / git-auth-integration-test` - Runs the `GitHub` and `Azure DevOps` API and authentication integration test suite for Serverless. For the configuration, see the [_integration-tests.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/_integration-tests.yaml) file.
- `push / upgrades / operator-upgrade-test` - Runs the upgrade integration test suite and verifies if the latest release can be successfully upgraded to the new (`main`) revision. For the configuration, see the [_upgrade-tests.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/_upgrade-tests.yaml) file.
- `push / upgrades / serverless-upgrade-test` - Runs the basic functionality integration and the `tracing`, `api-gateway`, and `cloud-event` contract compatibility integration test suite for Serverless in a k3d cluster after upgrading from the latest release to the actual revision to check if the Serverless component is working properly after the upgrade. For the configuration, see the [_upgrade-tests.yaml](https://github.com/kyma-project/serverless/blob/main/.github/workflows/_upgrade-tests.yaml) file.

## Smoke-Test Serverless Module on a Given Cluster

Follow these steps to verify that the serverless module works on your Kyma instance:
1. Clone this repository locally.
2. Point KUBECONFIG environment variable to the file containing kubeconfig configuration of your cluster.

```
export KUBECONFIG=<path-to-kubeconfig>
```

3. Check if the `Serverless` Custom Resource is in the Ready state using the following command:

```
kubectl get serverlesses.operator.kyma-project.io -n kyma-system
NAME      CONFIGURED   INSTALLED   GENERATION   AGE   STATE
default   True         True        3            27h   Ready
```

4. Run the tests using the following make targets in the root of the cloned repository:

```
make -C tests/serverless serverless-integration serverless-contract-tests
```