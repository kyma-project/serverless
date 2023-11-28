#!/usr/bin/env bash

echo "Checking status of github actions for serverless-manager"

REF_NAME="${1:-"main"}"
RAW_EXPECTED_SHA=$(git log "${REF_NAME}" --max-count 1 --format=format:%H)
REPOSITORY_ID="563346860"

STATUS_URL="https://api.github.com/repositories/${REPOSITORY_ID}/actions/workflows/gardener-integration.yaml/runs?head_sha=${RAW_EXPECTED_SHA}"
GET_STATUS_JQ_QUERY=".workflow_runs[0] | \"\(.status)-\(.conclusion)\""
GET_COUNT_JQ_QUERY=".total_count"

response=`curl -s ${STATUS_URL}`

count=`echo $response | jq -r "${GET_COUNT_JQ_QUERY}"`
if [[ "$count" == "0" ]]; then
  echo "No actions to verify"
else
  fullstatus=`echo $response |  jq -r "${GET_STATUS_JQ_QUERY}"`
  if [[ "$fullstatus" == "completed-success" ]]; then
    echo "All actions succeeded"
  else
    echo "Actions failed or pending - Check github actions status"
    exit 1
  fi
fi
