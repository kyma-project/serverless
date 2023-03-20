# Serverless Manager

## Overview

Serverless Manager allows deploying the [Serverless](https://kyma-project.io/docs/kyma/latest/01-overview/main-areas/serverless/) component on the Kyma cluster in compatibility with the [Lifecycle Manager](https://github.com/kyma-project/lifecycle-manager).

## Install

> **NOTE:** serverless-manager temporarily has a dependency to `kyma/cluster-essentials`


To install serverless-manager simply apply the following script:

```bash
kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/serverless-manager.yaml
```

To get Serverless installed, apply the sample Serverless CR:

```bash
kubectl apply -f config/samples/operator_v1alpha1_serverless_k3d.yaml
```

## Development

### Prerequisites

- Access to a k8s cluster
- [Go](https://go.dev/)
- [k3d](https://k3d.io/)
- [Docker](https://www.docker.com/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Kubebuilder](https://book.kubebuilder.io/)


## Manual installation using make targets

1. Clone the project.

    ```bash
    git clone https://github.com/kyma-project/serverless-manager.git && cd serverless-manager/
    ```

2. Set the `serverless-manager` image name.

    ```bash
    export IMG=<DOCKER_USERNAME>/custom-serverless-manager:0.0.1
    ```

3. Verify the compability.

    ```bash
    make test
    ```

4. Build and push the image to the registry.

    ```bash
    make module-image
    ```

5. Deploy.

    ```bash
    make deploy
    ```

### Test integration with lifecycle-manager on the k3d cluster

1. Clone the project.

    ```bash
    git clone https://github.com/kyma-project/serverless-manager.git && cd serverless-manager/
    ```

2. Build the manager locally and run it on the k3d cluster.

    ```bash
    make local-run
    ```

> **NOTE:** To clean up the k3d cluster, use the `make local-stop` make target.


## Using `serverless-manager`

- Create a Serverless instance.

    ```bash
    kubectl apply -f config/samples/operator_v1alpha1_serverless_k3d.yaml
    ```

- Delete a Serverless instance.

    ```bash
    kubectl delete -f config/samples/operator_v1alpha1_serverless_k3d.yaml
    ```

- Use external registry.

    The following example shows how you can modify the Serverless docker registry address using the `serverless.operator.kyma-project.io` CR:

    ```bash
    kubectl create secret generic my-secret \
        --namespace kyma-system \
        --from-literal username="<USERNAME>" \
        --from-literal password="<PASSWORD>" \
        --from-literal serverAddress="<SERVER_ADDRESS>" \
        --from-literal registryAddress="<REGISTRY_ADDRESS>"
    ```

    > **NOTE:** For DockerHub: 
    > - SERVER_ADDRESS is "https://index.docker.io/v1/",
    > - USERNAME and REGISTRY_ADDRESS must be identical.

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: operator.kyma-project.io/v1alpha1
    kind: Serverless
    metadata:
    name: serverless-sample
    spec:
        dockerRegistry:
            enableInternal: false
            secretName: my-secret
    EOF
    ```
## Testing Strategy

Each pull request to the repository triggers CI/CD jobs that verify serverless manager reconciliation logic and run integration tests of serverless module.

- `pre-serverless-manager-operator-build` - Compiling serverless manager code and pushing it's docker image.
- `pre-serverless-manager-operator-tests` - Testing serverless manager reconciliation code (Serverless CR CRUD operations).
- `pre-main-serverless-manager-verify` - Integration testing for serverless module installed by serverless-manager.
- `pull-serverless-module-build` - Bundling a module template manifest that allows testing it against lifecycle-manager manually. 

After pull request is merged a collection of CI/CD jobs are executed that:
 - re-builds serverless manager image
 - rebuilds serverless module and prepares module template manifest file that could be submitted to modular kyma
 - tests integration with lifecycle-manager
 
## Troubleshooting

- For MacBook M1 users

    Some parts of the scripts may not work because the Kyma CLI is not released for Apple Silicon users. To fix it [install Kyma CLI manually](https://github.com/kyma-project/cli#installation) and export the path to it.

    ```bash
    export KYMA=$(which kyma)
    ```

    > The example error may look like this: `Error: unsupported platform OS_TYPE: Darwin, OS_ARCH: arm64; to mitigate this problem set variable KYMA with the absolute path to kyma-cli binary compatible with your operating system and architecture.  Stop.`
