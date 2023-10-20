#!/bin/bash

source "collect_logs.sh"

kubectl proxy &
KUBECTL_PID=$!

APP_TEST_KUBECTL_PROXY_ENABLED=true go run ./cmd/main.go "$1"
EXIT_CODE=$?
kill $KUBECTL_PID

if [[ $EXIT_CODE -ne 0 ]]; then
  collect_logs
fi
exit ${EXIT_CODE}
