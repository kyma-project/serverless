# Serverless

## Status
![GitHub tag checks state](https://img.shields.io/github/checks-status/kyma-project/serverless-manager/main?label=serverless-operator&link=https%3A%2F%2Fgithub.com%2Fkyma-project%2Fserverless-manager%2Fcommits%2Fmain)
[![REUSE status](https://api.reuse.software/badge/github.com/kyma-project/serverless-manager)](https://api.reuse.software/info/github.com/kyma-project/serverless-manager)


## Overview

Serverless Operator allows deploying the [Serverless](https://kyma-project.io/docs/kyma/latest/01-overview/serverless/) component on the Kyma cluster in compatibility with [Lifecycle Manager](https://github.com/kyma-project/lifecycle-manager).

## Install

Create the `kyma-system` namespace:

```bash
kubectl create namespace kyma-system
```

Apply the following script to install Serverless Operator:

```bash
kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/serverless-operator.yaml
```

To get Serverless installed, apply the sample Serverless CR:

```bash
kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/default-serverless-cr.yaml
```

## Development

### Prerequisites

- Access to a Kubernetes (v1.24 or higher) cluster
- [Go](https://go.dev/)
- [k3d](https://k3d.io/)
- [Docker](https://www.docker.com/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Kubebuilder](https://book.kubebuilder.io/)


## Manual Installation Using Make Targets

1. Clone the project.

    ```bash
    git clone https://github.com/kyma-project/serverless-manager.git && cd serverless-manager/
    ```

2. Set the Serverless Operator image name.

    ```bash
    export IMG=<DOCKER_USERNAME>/custom-serverless-operator:0.0.1
    ```

3. Verify the compability.

    ```bash
    make test
    ```

4. Build and push the image to the registry.

    ```bash
    make module-image-release
    ```

5. Deploy Serverless Operator.

    ```bash
    make deploy
    ```

### Test Integration with Lifecycle Manager on the k3d Cluster

1. Clone the project.

    ```bash
    git clone https://github.com/kyma-project/serverless-manager.git && cd serverless-manager/
    ```

2. Build Serverless Operator locally and run it on the k3d cluster.

    ```bash
    make -C hack/local run-with-lifecycle-manager
    ```

> **NOTE:** To clean up the k3d cluster, use the `make -C hack/local stop` make target.


## Using Serverless Operator

- Create a Serverless instance.

    ```bash
    kubectl apply -f config/samples/default-serverless-cr.yaml
    ```

- Delete a Serverless instance.

    ```bash
    kubectl delete -f config/samples/default-serverless-cr.yaml
    ```

- Use external registry.

    The following example shows how you can modify the Serverless Docker registry address using the `serverless.operator.kyma-project.io` CR:

    ```bash
    kubectl create secret generic my-secret \
        --namespace kyma-system \
        --from-literal username="<USERNAME>" \
        --from-literal password="<PASSWORD>" \
        --from-literal serverAddress="<SERVER_ADDRESS>" \
        --from-literal registryAddress="<REGISTRY_ADDRESS>"
    ```

    > **NOTE:** For DockerHub: 
    <!-- markdown-link-check-disable-next-line -->
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