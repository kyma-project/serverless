#!/bin/bash

function get_dockerregistry_status () {
	local number=1
	while [[ $number -le 100 ]] ; do
		echo ">--> checking dockerregistry status #$number"
		local STATUS=$(kubectl get dockerregistry -n kyma-system default -o jsonpath='{.status.state}')
		echo "dockerregistry status: ${STATUS:='UNKNOWN'}"
		[[ "$STATUS" == "Ready" ]] && return 0
		sleep 5
        	((number = number + 1))
	done

	kubectl get all --all-namespaces
	exit 1
}

get_dockerregistry_status
