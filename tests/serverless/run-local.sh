#!/bin/bash

set -e -o pipefail

function cleanup() {
  kill $KUBECTL_PID || true
}

kubectl proxy &
KUBECTL_PID=$!

trap cleanup SIGINT SIGTERM EXIT

APP_TEST_KUBECTL_PROXY_ENABLED=true APP_TEST_CLEANUP=onSuccessOnly GODEBUG=fips140=only,tlsmlkem=0 go run ./cmd/main.go "$1"
EXIT_CODE=$?

kill $KUBECTL_PID || true

exit ${EXIT_CODE}
