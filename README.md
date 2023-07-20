# Serverless Operator

## Overview

Serverless Operator allows deploying the [Serverless](https://kyma-project.io/docs/kyma/latest/01-overview/serverless/) component on the Kyma cluster in compatibility with the [Lifecycle Manager](https://github.com/kyma-project/lifecycle-manager).

## Install

To install serverless-operator simply apply the following script:

```bash
kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/serverless-operator.yaml
```

To get Serverless installed, apply the sample Serverless CR:

```bash
kubectl apply -f config/samples/operator_v1alpha1_serverless.yaml
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

2. Set the `serverless-operator` image name.

    ```bash
    export IMG=<DOCKER_USERNAME>/custom-serverless-operator:0.0.1
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

2. Build the serverless operator locally and run it on the k3d cluster.

    ```bash
    make -C hack/local run-with-lifecycle-manager
    ```

> **NOTE:** To clean up the k3d cluster, use the `make -C hack/local stop` make target.


## Using `serverless-operator`

- Create a Serverless instance.

    ```bash
    kubectl apply -f config/samples/operator_v1alpha1_serverless.yaml
    ```

- Delete a Serverless instance.

    ```bash
    kubectl delete -f config/samples/operator_v1alpha1_serverless.yaml
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

Each pull request to the repository triggers CI/CD jobs that verify the serverless-operator reconciliation logic and run integration tests of the Serverless module.

- `pre-serverless-operator-operator-build` - Compiling the serverless-operator code and pushing its docker image.
- `pre-serverless-operator-operator-tests` - Testing the serverless-operator reconciliation code (Serverless CR CRUD operations).
- `pre-main-serverless-operator-verify` - Integration testing for the Serverless module installed by serverless-operator (**not using lifecycle-manager**).
- `pull-serverless-module-build` - Bundling a module template manifest that allows testing it against lifecycle-manager manually. 

After the pull request is merged, the following CI/CD jobs are executed:

 - `post-main-serverless-operator-verify` - Installs the Serverless module (**using lifecycle-manager**) and runs integration tests of Serverless.
 - `post-serverless-operator-build` - rebuilds  the Serverless module and the module template manifest file that can be submitted to modular Kyma.
 
## Troubleshooting

- For MacBook M1 users

    Some parts of the scripts may not work because the Kyma CLI is not released for Apple Silicon users. To fix it [install Kyma CLI manually](https://github.com/kyma-project/cli#installation) and export the path to it.

    ```bash
    export KYMA=$(which kyma)
    ```

    > The example error may look like this: `Error: unsupported platform OS_TYPE: Darwin, OS_ARCH: arm64; to mitigate this problem set variable KYMA with the absolute path to kyma-cli binary compatible with your operating system and architecture.  Stop.`
