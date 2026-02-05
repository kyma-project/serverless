# Configuring Logging

This document describes how to configure logging for Serverless components. The Serverless module supports dynamic log reconfiguration through the Serverless custom resource (CR).


> [!NOTE]
> - **logLevel** changes are applied dynamically without restarting the Serverless Pods
> - **logFormat** changes automatically restart the Serverless Pods to apply the new format

## Supported Log Levels

From the least to the most verbose: `fatal`, `panic`, `dpanic`, `error`, `warn`, `info` (default), `debug`.

## Supported Log Formats

- `json` - Structured JSON format (default)
- `console` - Human-readable console format

## Configure Serverless Controller

Update the log configuration by modifying the Serverless CR in the `kyma-system` namespace:

   ```bash
   # Change log level (updates ConfigMap without restarting Pods)
   kubectl patch serverless -n kyma-system default --type merge -p '{"spec":{"logLevel":"debug"}}'

   # Change log format (updates ConfigMap and restarts Pods)
   kubectl patch serverless -n kyma-system default --type merge -p '{"spec":{"logFormat":"console"}}'

   # Change both level and format
   kubectl patch serverless -n kyma-system default --type merge -p '{"spec":{"logLevel":"debug","logFormat":"console"}}'
   ```

Verify the change:

   ```bash
   kubectl logs -n kyma-system -l app.kubernetes.io/name=serverless
   ```

## Configuring Serverless Operator

Update the log configuration in the `serverless-operator-log-config` ConfigMap in the `kyma-system` namespace:

   ```bash
   # Change log level
   kubectl patch configmap serverless-operator-log-config -n kyma-system --type merge -p '{"data":{"log-config.yaml":"logLevel: debug"}}'

   # Change log level and format
   kubectl patch configmap serverless-operator-log-config -n kyma-system --type merge -p '{"data":{"log-config.yaml":"logLevel: debug\nlogFormat: console"}}'
   ```

Verify the change:

   ```bash
   kubectl logs -n kyma-system -l control-plane=operator
   ```

> [!NOTE]
> It is not possible to dynamically change the log format for the Serverless Operator. If you want to change it, update the ConfigMap and restart the Pods.
