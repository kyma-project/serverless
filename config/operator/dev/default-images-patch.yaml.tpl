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
            - name: IMAGE_FUNCTION_RUNTIME_NODEJS20
              value: ""
            - name: IMAGE_FUNCTION_RUNTIME_NODEJS20_FIPS
              value: ""
            - name: IMAGE_FUNCTION_RUNTIME_NODEJS22
              value: ""
            - name: IMAGE_FUNCTION_RUNTIME_NODEJS22_FIPS
              value: ""
            - name: IMAGE_FUNCTION_RUNTIME_NODEJS24
              value: ""
            - name: IMAGE_FUNCTION_RUNTIME_NODEJS24_FIPS
              value: ""
            - name: IMAGE_FUNCTION_RUNTIME_PYTHON312
              value: ""
            - name: IMAGE_FUNCTION_RUNTIME_PYTHON312_FIPS
              value: ""
            - name: IMAGE_FUNCTION_RUNTIME_PYTHON314
              value: ""
            - name: IMAGE_FUNCTION_RUNTIME_PYTHON314_FIPS
              value: ""
            - name: IMAGE_FUNCTION_BUILDLESS_CONTROLLER
              value: ""
            - name: IMAGE_FUNCTION_BUILDLESS_CONTROLLER_FIPS
              value: ""
            - name: IMAGE_FUNCTION_BUILDLESS_INIT
              value: ""
            - name: IMAGE_FUNCTION_BUILDLESS_INIT_FIPS
              value: ""
            - name: IMAGE_KANIKO_EXECUTOR
              value: ""
            - name: IMAGE_FUNCTION_CONTROLLER
              value: ""
            - name: IMAGE_FUNCTION_BUILD_INIT
              value: ""
            - name: IMAGE_REGISTRY_INIT
              value: ""
            - name: IMAGE_REGISTRY
              value: ""
