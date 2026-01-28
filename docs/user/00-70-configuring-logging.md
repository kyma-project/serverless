# Configuring Logging

This document describes how to configure logging for Serverless components. The Serverless module supports dynamic log reconfiguration through Kubernetes ConfigMaps.


> [NOTE]
> It is not possible to dynamically change the log format without restarting the Pod. If you want your format to persist use `kubectl rollout restart deployment <deployment-name>` after changing the format in the ConfigMap for a zero downtime restart.

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

## Configuring Serverless Operator

Update the log configuration in the `serverless-operator-log-config` ConfigMap in the `kyma-system` namespace:

   ```bash
   # Change log level (applied immediately without restart)
   kubectl patch configmap serverless-operator-log-config -n kyma-system --type merge -p '{"data":{"log-config.yaml":"logLevel: debug\nlogFormat: json"}}'

   # Change log format (triggers an automatic Pod restart)
   kubectl patch configmap serverless-operator-log-config -n kyma-system --type merge -p '{"data":{"log-config.yaml":"logLevel: debug\nlogFormat: console"}}'
   ```

[!TIP]

   ```bash
   kubectl logs -n kyma-system -l control-plane=operator
   ```
