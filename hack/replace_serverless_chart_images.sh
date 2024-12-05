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


# temporary loop - finally we will only do the replacement in one of the serverless
SERVERLESSES=('serverless' 'buildless-serverless')
for SERVERLESS in "${SERVERLESSES[@]}"; do
VALUES_FILE=${PROJECT_ROOT}/config/${SERVERLESS}/values.yaml

if [[ ${PURPOSE} == "local" ]]; then
  echo "Changing container registry (${SERVERLESS})"
  yq -i '.global.containerRegistry.path="k3d-kyma-registry.localhost:5000"' "${VALUES_FILE}"
fi

IMAGES_SELECTOR=".global.images[] | select(key == \"function_*\") ${MAIN_ONLY_SELECTOR}"
yq --inplace "(${IMAGES_SELECTOR} | .directory) = \"${IMG_DIRECTORY}\"" ${VALUES_FILE}
yq --inplace "(${IMAGES_SELECTOR} | .version) = \"${IMG_VERSION}\"" ${VALUES_FILE}
echo "==== Local Changes (${SERVERLESS}) ===="
yq '.global.images' "${VALUES_FILE}"
yq '.global.containerRegistry' "${VALUES_FILE}"
echo "==== End of Local Changes (${SERVERLESS}) ===="
done
