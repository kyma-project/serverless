#!/usr/bin/env bash

echo "Checking status of github actions for serverless-manager"

REF_NAME="${1:-"main"}"
RAW_EXPECTED_SHA=$(git log "${REF_NAME}" --max-count 1 --format=format:%H)
STATUS_URL="https://api.github.com/repos/kyma-project/serverless-manager/actions/workflows/gardener-integration.yaml/runs"
JQ_QUERY="[.workflow_runs[] | select(.head_sha | test(\"${RAW_EXPECTED_SHA}\"))][0] | \"\(.status)-\(.conclusion)\""
fullstatus=`curl -s ${STATUS_URL} |  jq -r "${JQ_QUERY}"`

echo $fullstatus

if [[ "$fullstatus" == "completed-success" ]]; then
  echo "All actions succeeded"
else
  echo "Actions failed or pending - Check github actions status"
  exit 1
fi