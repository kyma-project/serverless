# Serverless

## Status
![GitHub tag checks state](https://img.shields.io/github/checks-status/kyma-project/serverless/main?label=serverless-operator&link=https%3A%2F%2Fgithub.com%2Fkyma-project%2Fserverless%2Fcommits%2Fmain)
<!-- markdown-link-check-disable-next-line -->
[![REUSE status](https://api.reuse.software/badge/github.com/kyma-project/serverless)](https://api.reuse.software/info/github.com/kyma-project/serverless)


## Overview

Serverless Operator allows deploying the [Serverless](https://kyma-project.io/docs/kyma/latest/01-overview/serverless/) component in the Kyma cluster in compatibility with [Lifecycle Manager](https://github.com/kyma-project/lifecycle-manager).

### Architecture Diagram
![Architecture](./architecture.svg)

## Install

Create the `kyma-system` namespace:

```bash
kubectl create namespace kyma-system
```

Apply the following script to install Serverless Operator:

```bash
kubectl apply -f https://github.com/kyma-project/serverless/releases/latest/download/serverless-operator.yaml
```

To get Serverless installed, apply the sample Serverless CR:

```bash
kubectl apply -f https://github.com/kyma-project/serverless/releases/latest/download/default-serverless-cr.yaml
```

## Development

### Prerequisites

- Access to a Kubernetes (v1.24 or higher) cluster
- [Go](https://go.dev/)
- [k3d](https://k3d.io/)
- [Docker](https://www.docker.com/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Kubebuilder](https://book.kubebuilder.io/)
- [yq](https://mikefarah.gitbook.io/yq)


## Installation in the k3d Cluster Using Make Targets

1. Clone the project.

    ```bash
    git clone https://github.com/kyma-project/serverless.git && cd serverless/
    ```

2. Create a new k3d cluster and run Serverless from the main branch:

    ```bash
    make run-main
    ```

> **NOTE:** To clean up the k3d cluster, use the `make delete-k3d` make target.

> **NOTE:** If you have k3d already running, you can use the `install-*` targets to install Serverless in different flavors.

## Using Serverless Operator

- Create a Serverless instance.

    ```bash
    kubectl apply -f config/samples/legacy-serverless-cr.yaml
    ```

- Delete a Serverless instance.

    ```bash
    kubectl delete -f config/samples/legacy-serverless-cr.yaml
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

## Contributing

See the [Contributing Rules](CONTRIBUTING.md).

## Code of Conduct

See the [Code of Conduct](CODE_OF_CONDUCT.md) document.

## Licensing

See the [license](./LICENSE) file.
