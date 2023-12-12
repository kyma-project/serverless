#!/usr/bin/env bash

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
		sleep 5
        	((number = number + 1))
	done

	exit 1
}

verify_github_jobs_status