# buildless-serverless

The buildless-serverless controller manages the lifecycle of Kubernetes Functions. It runs user code directly in containers using pre-built runtime images.

## How It Works

For each `Function` CR, the controller:
1. Validates the Function spec
2. Injects the function source code into a runtime container (via init container for git-sourced Functions, or directly for inline Functions)
3. Creates/updates a `Deployment` and `Service` for the Function

Supported runtimes are listed in the [`components/runtimes`](../runtimes) directory.

## Prerequisites

- [Go](https://go.dev/)

## Development

### Running Locally

```bash
make run
```

### Environment Variables

| Variable | Description | Default value |
|---|---|---|
| **APP_FUNCTION_CONFIG_PATH** | Path to the function configuration YAML file | `hack/function-config.yaml` |
| **APP_LOG_CONFIG_PATH** | Path to the log configuration YAML file | `hack/log-config.yaml` |
| **APP_KYMA_FIPS_MODE_ENABLED** | Enables FIPS 140 exclusive mode | `false` |
