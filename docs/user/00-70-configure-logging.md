# Configure Logging

This document describes how to configure logging for Serverless components. The Serverless module supports dynamic log reconfiguration through Kubernetes ConfigMaps.

> **Note:** 
> - **Log level** changes are applied dynamically without pod restart
> - **Log format** changes trigger an automatic pod restart to apply the new format

## Supported Log Levels

From least to most verbose: `fatal`, `panic`, `dpanic`, `error`, `warn`, `info` (default), `debug`

## Supported Log Formats

- `json` - Structured JSON format (default)
- `console` - Human-readable console format

## Configure Serverless Controller

Update the log configuration in the `serverless-log-config` ConfigMap in the `kyma-system` namespace:

```bash
# Change log level (applies immediately without restart)
kubectl patch configmap serverless-log-config -n kyma-system --type merge -p '{"data":{"log-config.yaml":"logLevel: debug\nlogFormat: json"}}'

# Change log format (triggers automatic pod restart)
kubectl patch configmap serverless-log-config -n kyma-system --type merge -p '{"data":{"log-config.yaml":"logLevel: debug\nlogFormat: console"}}'

# Change both level and format
kubectl patch configmap serverless-log-config -n kyma-system --type merge -p '{"data":{"log-config.yaml":"logLevel: info\nlogFormat: json"}}'
```

Verify the change:

```bash
kubectl logs -n kyma-system -l app.kubernetes.io/name=serverless
```
If you want to see logs from previous instance (before format change), use the `--previous` flag.


### Keep in mind

If you change the log format the pod will gracefully restart, so you might need to wait a moment before checking the logs. This is expected behavior as updating controller log format is not possible without a restart.

## Configure Serverless Operator

Update the log configuration in the `serverless-operator-log-config` ConfigMap in the `kyma-system` namespace:

```bash
# Change log level (applies immediately without restart)
kubectl patch configmap serverless-operator-log-config -n kyma-system --type merge -p '{"data":{"log-config.yaml":"logLevel: debug\nlogFormat: json"}}'

# Change log format (triggers automatic pod restart)
kubectl patch configmap serverless-operator-log-config -n kyma-system --type merge -p '{"data":{"log-config.yaml":"logLevel: debug\nlogFormat: console"}}'
```

Verify the change:

```bash
kubectl logs -n kyma-system -l control-plane=operator
```

