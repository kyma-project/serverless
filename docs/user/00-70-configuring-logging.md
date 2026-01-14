# Configuring Logging

This document describes how to configure logging for Serverless components. The Serverless module supports dynamic log reconfiguration through Kubernetes ConfigMaps.

[!NOTE]
> - **Log level** changes are applied dynamically without a Pod restart
> - **Log format** changes trigger an automatic Pod restart to apply the new format

## Supported Log Levels

From the least to the most verbose: `fatal`, `panic`, `dpanic`, `error`, `warn`, `info` (default), `debug`.

## Supported Log Formats

- `json` - Structured JSON format (default)
- `console` - Human-readable console format

## Configure Serverless Controller

Update the log configuration in the `serverless-log-config` ConfigMap in the `kyma-system` namespace:

   ```bash
   # Change log level (applied immediately without restart)
   kubectl patch configmap serverless-log-config -n kyma-system --type merge -p '{"data":{"log-config.yaml":"logLevel: debug\nlogFormat: json"}}'

   # Change log format (triggers an automatic Pod restart)
   kubectl patch configmap serverless-log-config -n kyma-system --type merge -p '{"data":{"log-config.yaml":"logLevel: debug\nlogFormat: console"}}'

   # Change both, level and format
   kubectl patch configmap serverless-log-config -n kyma-system --type merge -p '{"data":{"log-config.yaml":"logLevel: info\nlogFormat: json"}}'
   ```

Verify the change:

   ```bash
   kubectl logs -n kyma-system -l app.kubernetes.io/name=serverless
   ```
If you want to see logs a from previous instance (before the format change), use the `--previous` flag.

> [NOTE]
> If you change the log format, the Pod restarts gracefully, so you might need to wait a moment before checking the logs. This is expected behavior as updating the log format controller is not possible without a restart.

## Configuring Serverless Operator

Update the log configuration in the `serverless-operator-log-config` ConfigMap in the `kyma-system` namespace:

   ```bash
   # Change log level (applied immediately without restart)
   kubectl patch configmap serverless-operator-log-config -n kyma-system --type merge -p '{"data":{"log-config.yaml":"logLevel: debug\nlogFormat: json"}}'

   # Change log format (triggers an automatic Pod restart)
   kubectl patch configmap serverless-operator-log-config -n kyma-system --type merge -p '{"data":{"log-config.yaml":"logLevel: debug\nlogFormat: console"}}'
   ```

Verify the change:

   ```bash
   kubectl logs -n kyma-system -l control-plane=operator
   ```

