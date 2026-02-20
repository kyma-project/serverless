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
            - name: KYMA_FIPS_MODE_ENABLED
              value: "false"
