#!/bin/bash

source "collect_logs.sh"

kubectl proxy &
KUBECTL_PID=$!

APP_TEST_KUBECTL_PROXY_ENABLED=true APP_TEST_CLEANUP=onSuccessOnly INCLUDE_GITOPS_TEST="$INCLUDE_GITOPS_TEST" go run ./cmd/main.go "$1"
EXIT_CODE=$?
kill $KUBECTL_PID

if [[ $EXIT_CODE -ne 0 ]]; then
  echo "test failed"
  collect_logs
fi
exit ${EXIT_CODE}
