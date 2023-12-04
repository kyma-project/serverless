#!/usr/bin/env bash

valid=true

function check_image() {
    local version=$(jq '.version' <<< $1)
    local name=$(jq '.name' <<< $1)

    if [[ $version == *PR-* || $version == null ]]; then
      echo "Invalid version: ${version} in ${name}"
      valid=false
    fi
}

versionNameJson="$(yq -o json '.global.images[] | [{"version": .version, "name": .name}][]' ./config/serverless/values.yaml)"
for versionName in $(jq --compact-output <<< $versionNameJson); do
  check_image "${versionName}"
done

if [ $valid == false ]; then
  exit 1
fi
echo "All versions are valid"
