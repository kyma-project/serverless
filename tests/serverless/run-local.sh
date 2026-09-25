#!/bin/bash

set -e -o pipefail

function cleanup() {
  kill $KUBECTL_PID || true
}

kubectl apply -f ../../tests/fixtures/serverless-cr-integration-test.yaml

# Wait for the operator to reconcile the configmap with the updated requeue duration
SECONDS_WAITED=0
until kubectl get configmap serverless-config -n kyma-system -o jsonpath='{.data.function-config\.yaml}' 2>/dev/null | grep -q 'functionReadyRequeueDuration: "30s"'; do
  if [ $SECONDS_WAITED -ge 60 ]; then
    echo "Timed out waiting for serverless-config to reflect functionReadyRequeueDuration=30s"
    exit 1
  fi
  sleep 2
  SECONDS_WAITED=$((SECONDS_WAITED + 2))
done

# Restart controller to load the updated configmap
kubectl rollout restart deployment/serverless-ctrl-mngr -n kyma-system
kubectl rollout status deployment/serverless-ctrl-mngr -n kyma-system --timeout=60s

kubectl proxy &
KUBECTL_PID=$!

trap cleanup SIGINT SIGTERM EXIT

APP_TEST_KUBECTL_PROXY_ENABLED=true APP_TEST_CLEANUP=onSuccessOnly GODEBUG=fips140=only,tlsmlkem=0 go run ./cmd/main.go "$1"
EXIT_CODE=$?

kill $KUBECTL_PID || true

exit ${EXIT_CODE}
