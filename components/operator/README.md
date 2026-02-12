# Function Operator

The Function Operator is a Kubernetes controller that install Serverless.

## Prerequisites

The Function Operator requires the following components to be installed:

- [Go](https://go.dev/)

## Development

### Installing Locally Built Serverless Operator Image on k3d With Serverless
Prerequisite:
- installed Operator on a cluster

To install `serverless-manager` from local sources on the k3d cluster run:

```bash
make install-operator-k3d
```

This target:

- builds local image with `serverless-operator`
- tags local the `serverless-operator` image with its hash to be sure the deployment is updated
- uploads local image to k3d
- updates the image in the `serverless-operator` deployment

### Running/Debugging `serverless-operator` Locally

To develop the component as a normal binary, set up environment variables.

### Environment Variables

#### The Function Operator Uses These Environment Variables:

| Variable                   | Description                                                                                                         | Default value   |
|----------------------------|---------------------------------------------------------------------------------------------------------------------|-----------------|
| **CHART_PATH**             | Location of the Serverless chart                                                                                        | `/buildless-module-chart` |
| **SERVERLESS_MANAGER_UID** | Unique ID of operator instance. Used to mark created resources to distinguish which version of the Operator created it. | ``              |
