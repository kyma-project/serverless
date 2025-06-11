# Overview

Serverless exposes metrics for collectors like Prometheus. For local development, you can use a local instance of Prometheus.
This document describes how to set up such an environment.

## Run Serverless

Navigate to the `components/buildless-serverless`) directory and run:

```bash
make run
```

## Test Serverless Metrics

To test the metrics, run the following command:

```bash
curl http://localhost:8080/metrics | grep serverless
```

You should see an output similar to the following:

```
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100 18524    0 18524    0     0  10.7M      0 --:--:-- --:--:-- --:--:-- 17.6M
# HELP serverless_resources_processed_total Total number of resources processed for the first time, partitioned by user runtime
# TYPE serverless_resources_processed_total counter
serverless_resources_processed_total{runtime="makapaka"} 1
serverless_resources_processed_total{runtime="nodejs20"} 3
# HELP serverless_version_info Static metric with Serverless version info
# TYPE serverless_version_info gauge
serverless_version_info{version="7.8.7"} 1
```

## Deploy Prometheus

Go to the `examples/prometheus-local` directory and apply the following manifests:

```bash
kubectl apply -f metrics-ns.yaml
kubectl apply -f prometheus-config.yaml
kubectl apply -f prometheus.yaml
kubectl apply -f prometheus-svc.yaml
```

To access Prometheus in your browser, expose the service locally:

```bash
kubectl port-forward svc/prometheus 9090:9090 -n metrics
```

## Use Prometheus

Open your browser and go to 
`http://localhost:9090/`

You can use the following query to inspect your metrics:

```
serverless_resources_processed_total
```

To list all available metrics along with their descriptions, you can also run:

```bash
curl http://localhost:8080/metrics | grep HELP
```
