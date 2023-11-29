#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
ROOT_DIR="${SCRIPT_DIR}/../../.."

LOCAL_REGISTRY=${1?"Registry adress missing"}


replace_values () {
  yq -i ".global.containerRegistry.path=\"${LOCAL_REGISTRY}\"" ${ROOT_DIR}/config/serverless/values.yaml
}
