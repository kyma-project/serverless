# Function Operator

The Function Operator is a Kubernetes controller that install Serverless.

## Prerequisites

The Function Operator requires the following components to be installed:

- [Go](https://go.dev/)

## Development

### Installing locally build serverless operator image on k3d with serverless
Prerequisite:
- installed operator on cluster

To install serverless-manager from local sources on k3d cluster run:

```bash
make install-operator-k3d
```


This target does:

- build local image with serverless-operator
- tag local serverless-operator image with its hash to be sure to deployment be updated
- upload local image to k3d
- update image in serverless-operator deployment

### Running/Debugging serverless-operator locally

To develop the component as normal binary, set up environment variables.

### Environment Variables

#### The Function Operator Uses These Environment Variables:

| Variable                   | Description                                                                                                         | Default value   |
|----------------------------|---------------------------------------------------------------------------------------------------------------------|-----------------|
| **CHART_PATH**             | Location of serverless chart                                                                                        | `/module-chart` |
| **SERVERLESS_MANAGER_UID** | Unique ID of operator instance. Used to mark created resources to distinguish which version of operator created it. | ``              |
