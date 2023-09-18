#!/usr/bin/env bash

function check_image() {
    local version=$1
    if [[ $version == PR-* || $version == null ]]; then
      echo "Invalid version: $version"
      exit 1
    fi
}

data=$(curl -s "https://raw.githubusercontent.com/kyma-project/kyma/release-2.18/resources/serverless/values.yaml" | yq -r '.global.images.function_controller.version, .global.images.function_webhook.version, .global.images.function_registry_gc.version')

images=($data)

check_image "${images[0]}"
check_image "${images[1]}"
check_image "${images[2]}"

echo "All versions are valid"
