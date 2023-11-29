#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

IMG_DIRECTORY=${1?"Directory missing"}
IMG_VERSION=${2?"PULL_BASE_REF Missing"}


replace_values () {
  yq -i ".global.images.${1}.directory=\"${IMG_DIRECTORY}\""  ../../../../config/serverless/values.yaml
  yq -i ".global.images.${1}.version=\"${IMG_VERSION}\"" ../../../../config/serverless/values.yaml
}


replace_values "function_controller"
replace_values "function_webhook"
replace_values "function_build_init"
replace_values "function_registry_gc"
replace_values "function_runtime_nodejs16"  #Joby sie nie odpalaja bez zmian do runtime'ow
replace_values "function_runtime_nodejs18"
replace_values "function_runtime_python39"

