# Overview

Serverless exposes metrics for collectors like Prometheus. For local development, you can use a local instance of Prometheus.
This document describes how to set up such an environment.

## Serverless

1. Run Operator with Serverless by running the following command (from the root directory of the project):


   ```bash
   make run-buildless-main
   ```

1. To test the Serverless metrics, run the following commands:

    
   ```bash
   kubectl port-forward -n kyma-system services/serverless-controller-manager 8070:8080
   ```

   ```bash
   curl http://localhost:8070/metrics
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

## Prometheus

1. Apply the following manifests to deploy Prometheus:

Create a namespace for monitoring components:

   ```bash
   kubectl apply -f - <<EOF
   apiVersion: v1
   kind: Namespace
   metadata:
     name: monitoring
   EOF
   ```

Prepare Prometheus configuration:

   ```bash
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
             - targets: ['serverless-controller-manager.kyma-system.svc.cluster.local:8080']
   EOF
   ```
   
Deploy Prometheus:
    
   ```bash
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
   ```
   
Create a service to expose Prometheus:

   ```bash
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
   kubectl port-forward svc/prometheus 9090:9090 -n monitoring
   ```

1. To use Prometheus, open your browser and go to `http://localhost:9090/`.


- You can use the following query to inspect your metrics:

   ```
   serverless_resources_processed_total
   ```

- To list all metrics from the Serverless in the Prometheus UI, you can use the following query:

   ```
   {job="buildless"}
   ```
  
- You can also check correctness of the Prometheus scrape configuration in the Prometheus UI by navigating to:

   ```
   Status -> Target health
   ```

- To list all available metrics along with their descriptions, run:

   ```bash
   curl http://localhost:8070/metrics | grep HELP
   ```

## Kubernetes State Metrics

1. To use metrics with kubernetes state, you need to install the `kube-state-metrics` component.

- You can do this by running:

   ```bash
   helm install kube-state-metrics kube-state-metrics --repo https://prometheus-community.github.io/helm-charts --namespace monitoring
   ```
  
And check the available metrics by running:

   ```bash
   kubectl port-forward svc/kube-state-metrics -n monitoring 8060:8080
   ```

   ```bash
   curl http://localhost:8060/metrics
   ```
   
1. Apply the following manifests to see these metrics in Prometheus:

Create RBACs for the Prometheus auto-discovery:

   ```bash
   kubectl apply -f - <<EOF
   apiVersion: rbac.authorization.k8s.io/v1
   kind: ClusterRole
   metadata:
     name: prometheus-discovery
   rules:
     - apiGroups: [""]
       resources: ["nodes", "nodes/metrics", "services", "endpoints", "pods", "namespaces"]
       verbs: ["get", "list", "watch"]
   
   ---
   apiVersion: rbac.authorization.k8s.io/v1
   kind: ClusterRoleBinding
   metadata:
     name: prometheus-discovery-binding
   roleRef:
     apiGroup: rbac.authorization.k8s.io
     kind: ClusterRole
     name: prometheus-discovery
   subjects:
     - kind: ServiceAccount
       name: default
       namespace: monitoring
   EOF
   ```

Add scrape configuration for kube-state-metrics to the Prometheus configuration:

   ```bash
   kubectl edit cm -n monitoring prometheus-config
   ```
   
by adding the following to the `scrape_configs` section:

   ```yaml
      - job_name: 'kube-state-metrics-custom'
        metrics_path: /metrics
        kubernetes_sd_configs:
          - role: endpoints
            namespaces:
              names:
                - monitoring
        relabel_configs:
          - source_labels: [__meta_kubernetes_service_name]
            regex: kube-state-metrics
            action: keep
          - source_labels: [__meta_kubernetes_endpoint_port_name]
            regex: http
            action: keep
   ```
   
After editing the configuration, you need to restart the Prometheus deployment to apply the changes:

   ```bash
   kubectl rollout restart deployment -n monitoring prometheus
   ```

1. For custom kubernetes metrics, you can add your own definitions.

Create file `ksm-crd-values.yaml`:

    ```yaml
    customResourceState:
      enabled: true
      config:
        kind: CustomResourceStateMetrics
        spec:
          resources:
            - groupVersionKind:
                group: operator.kyma-project.io
                version: v1alpha1
                kind: Serverless
              labelsFromPath:
                name: ["metadata", "name"]
                namespace: ["metadata", "namespace"]
              metrics:
                - name: serverless_status
                  help: "Status of Serverless CR"
                  each:
                    type: StateSet
                    stateSet:
                      labelName: state
                      path: ["status", "state"]
                      list: [Ready, Processing, Error, Deleting, Warning]
                - name: serverless_condition
                  help: "Condition of Serverless"
                  each:
                    type: Gauge
                    gauge:
                      path: ["status", "conditions"]
                      labelsFromPath:
                        type: ["type"]
                        reason: ["reason"]
                      valueFrom: ["status"]
    rbac:
      extraRules:
        - apiGroups: ["operator.kyma-project.io"]
          resources: ["serverlesses"]
          verbs: ["list", "watch"]
   ``` 

Then upgrade the `kube-state-metrics`:

   ```bash
   helm upgrade kube-state-metrics prometheus-community/kube-state-metrics -n monitoring -f ksm-crd-values.yaml
   ```
   
Now in the Prometheus UI you can query your custom metrics:

   ```
   kube_customresource_serverless_status
   ```

   ```
   kube_customresource_serverless_condition
   ```