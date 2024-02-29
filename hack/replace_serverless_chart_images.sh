#!/bin/bash

# if you only need replace images with version set to "main" specify "main-only" argument
REPLACE_SCOPE=$1

REQUIRED_ENV_VARIABLES=('IMG_DIRECTORY' 'IMG_VERSION' 'PROJECT_ROOT')
for VAR in "${REQUIRED_ENV_VARIABLES[@]}"; do
  if [ -z "${!VAR}" ]; then
    echo "${VAR} is undefined"
    exit 1
  fi
done

MAIN_ONLY_SELECTOR=""
if [[ ${REPLACE_SCOPE} == "main-only" ]]; then
  MAIN_ONLY_SELECTOR="| select(.version == \"main\")"
fi

IMAGES_SELECTOR=".global.images[] | select(key == \"function_*\") ${MAIN_ONLY_SELECTOR}"
VALUES_FILE=${PROJECT_ROOT}/config/serverless/values.yaml

yq -i "(${IMAGES_SELECTOR} | .directory) = \"${IMG_DIRECTORY}\"" ${VALUES_FILE}
yq -i "(${IMAGES_SELECTOR} | .version) = \"${IMG_VERSION}\"" ${VALUES_FILE}
echo "==== Local Changes ===="
yq '.global.images' ${VALUES_FILE}
echo "==== End of Local Changes ===="
