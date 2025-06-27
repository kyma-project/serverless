# Overview

Serverless exposes metrics for collectors like Prometheus. For local development, you can use a local instance of Prometheus.
This document describes how to set up such an environment.

## Procedure

1. Run Serverless by navigating to the `components/buildless-serverless` directory and running the following command:


   ```bash
   make run
   ```

1. To test the Serverless metrics, run the following command:


   ```bash
   curl http://localhost:8080/metrics
   ```

You should see an output similar to the following:

   ```bash
   # HELP certwatcher_read_certificate_errors_total Total number of certificate read errors
   # TYPE certwatcher_read_certificate_errors_total counter
   certwatcher_read_certificate_errors_total 0
   # HELP certwatcher_read_certificate_total Total number of certificate reads
   # TYPE certwatcher_read_certificate_total counter
   certwatcher_read_certificate_total 0
   ...
   ```

1. Go to the `examples/prometheus-local` directory and apply the following manifests to deploy Prometheus:


   ```bash
   kubectl apply -f - <<EOF
   apiVersion: v1
   kind: Namespace
   metadata:
     name: monitoring
   EOF
   
   kubectl apply -f - <<EOF
   apiVersion: v1
   kind: ConfigMap
   metadata:
     name: prometheus-config
     namespace: monitoring
   data:
     prometheus.yml: |
       global:
         scrape_interval: 15s
       scrape_configs:
         - job_name: 'buildless'
           static_configs:
             - targets: ['buildless-serverless-controller-manager.kyma-system.svc.cluster.local:8080']
         - job_name: 'local-controller'
           static_configs:
             - targets: ['host.k3d.internal:8080']
   EOF
   
   kubectl apply -f - <<EOF
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: prometheus
     namespace: monitoring
   spec:
     replicas: 1
     selector:
       matchLabels:
         app: prometheus
     template:
       metadata:
         labels:
           app: prometheus
           networking.kyma-project.io/metrics-scraping: allowed
       spec:
         containers:
           - name: prometheus
             image: prom/prometheus
             args:
               - '--config.file=/etc/prometheus/prometheus.yml'
             ports:
               - containerPort: 9090
             volumeMounts:
               - name: prometheus-config-volume
                 mountPath: /etc/prometheus
         volumes:
           - name: prometheus-config-volume
             configMap:
               name: prometheus-config
   EOF
   
   kubectl apply -f - <<EOF
   apiVersion: v1
   kind: Service
   metadata:
     name: prometheus
     namespace: monitoring
   spec:
     type: NodePort
     ports:
       - port: 9090
         targetPort: 9090
         nodePort: 30090
     selector:
       app: prometheus
   EOF
   ```

1. To access Prometheus in your browser, expose the service locally:

   ```bash
   kubectl port-forward svc/prometheus 9090:9090 -n metrics
   ```

1. To use Prometheus, open your browser and go to `http://localhost:9090/`.


- You can use the following query to inspect your metrics:

   ```bash
   serverless_resources_processed_total
   ```

- To list all available metrics along with their descriptions, run:

   ```bash
   curl http://localhost:8080/metrics | grep HELP
   ```
