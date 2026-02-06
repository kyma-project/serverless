apiVersion: apps/v1
kind: Deployment
metadata:
  name: serverless-operator
  namespace: kyma-system
spec:
  template:
    spec:
      containers:
        - name: manager
          env:
            - name: IMAGE_FUNCTION_CONTROLLER
              value: ""
            - name: IMAGE_FUNCTION_BUILD_INIT
              value: ""
            - name: IMAGE_REGISTRY_INIT
              value: ""
            - name: IMAGE_FUNCTION_RUNTIME_NODEJS20
              value: ""
            - name: IMAGE_FUNCTION_RUNTIME_NODEJS22
              value: ""
            - name: IMAGE_FUNCTION_RUNTIME_NODEJS24
              value: ""
            - name: IMAGE_FUNCTION_RUNTIME_PYTHON312
              value: ""
            - name: IMAGE_KANIKO_EXECUTOR
              value: ""
            - name: IMAGE_REGISTRY
              value: ""
            - name: IMAGE_FUNCTION_BUILDLESS_CONTROLLER
              value: ""
            - name: IMAGE_FUNCTION_BUILDLESS_INIT
              value: ""
