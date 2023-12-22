#!/usr/bin/env bash

### Verify post-submit prow jobs status
#
# Optional input args:
#   - REF_NAME - branch/tag/commit
# Return status:
#   - return 0 - if status is "success"
#   - return 1 - if status is "failure" or after timeout (~25min)

# wait until Prow trigger pipelines
sleep 10

echo "Checking status of POST Jobs for Serverless"

REF_NAME="${1:-"main"}"
STATUS_URL="https://api.github.com/repos/kyma-project/serverless/commits/${REF_NAME}/status"

function verify_github_jobs_status () {
	local number=1
	while [[ $number -le 100 ]] ; do
		echo ">--> checking serverless job status #$number"
		local STATUS=`curl -L -H "Accept: application/vnd.github+json" -H "X-GitHub-Api-Version: 2022-11-28" ${STATUS_URL} | jq -r .state `
		echo "jobs status: ${STATUS:='UNKNOWN'}"
		[[ "$STATUS" == "success" ]] && return 0
		[[ "$STATUS" == "failure" ]] && return 1
		sleep 15
        	((number = number + 1))
	done

	exit 1
}

verify_github_jobs_status