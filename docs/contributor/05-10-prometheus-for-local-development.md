# Overview

Serverless exposes metrics for collectors like Prometheus. For local development, you can use a local instance of Prometheus.
This document describes how to set up such an environment.

## Get Serverless Metrics

1. Run Operator with Serverless by using the following command from the root directory of the project:


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

## Deploy Prometheus


1. Create a namespace for monitoring components:

   ```bash
   kubectl apply -f - <<EOF
   apiVersion: v1
   kind: Namespace
   metadata:
     name: monitoring
   EOF
   ```

2. Prepare Prometheus configuration:

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
   
3. Deploy Prometheus:
    
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
   
4. Create a service to expose Prometheus:

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

5. To access Prometheus in your browser, expose the service locally:

   ```bash
   kubectl port-forward svc/prometheus 9090:9090 -n monitoring
   ```

6. Open your browser and go to `http://localhost:9090/`.


- Use the following query to inspect your metrics:

   ```
   serverless_resources_processed_total
   ```

- To list all metrics from Serverless in the Prometheus UI, use the following query:

   ```
   {job="buildless"}
   ```
  
- To check the correctness of the Prometheus scrape configuration in the Prometheus UI, go to **Status -> Target health**


- To list all available metrics along with their descriptions, run:

   ```bash
   curl http://localhost:8070/metrics | grep HELP
   ```

## Use Serverless Metrics with kube-state-metrics

1. Install the `kube-state-metrics` component:

   ```bash
   helm install kube-state-metrics kube-state-metrics --repo https://prometheus-community.github.io/helm-charts --namespace monitoring
   ```
  
2. Check the available metrics:

   ```bash
   kubectl port-forward svc/kube-state-metrics -n monitoring 8060:8080
   ```

   ```bash
   curl http://localhost:8060/metrics
   ```

3. Create RBACs for the Prometheus auto-discovery:

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

4. Add scrape configuration for `kube-state-metrics` to the Prometheus configuration:

   ```bash
   kubectl edit cm -n monitoring prometheus-config
   ```
   
5. Add the following to the `scrape_configs` section:

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
   
6. Restart the Prometheus deployment to apply the changes:

   ```bash
   kubectl rollout restart deployment -n monitoring prometheus
   ```

### Adding Definitions for Custom Kubernetes Metrics

1. Create the ksm-crd-values.yaml file:

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

2. Upgrade `kube-state-metrics`:

   ```bash
   helm upgrade kube-state-metrics prometheus-community/kube-state-metrics -n monitoring -f ksm-crd-values.yaml
   ```
   
Now, in the Prometheus UI, you can query your custom metrics:

   ```
   kube_customresource_serverless_status
   ```

   ```
   kube_customresource_serverless_condition
   ```