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
  MAIN_ONLY_SELECTOR="| select(. == \"*:main$\")"
fi


# temporary loop - finally we will only do the replacement in one of the serverless
SERVERLESSES=('serverless' 'buildless-serverless')
for SERVERLESS in "${SERVERLESSES[@]}"; do
VALUES_FILE=${PROJECT_ROOT}/config/${SERVERLESS}/values.yaml
echo "processing ${VALUES_FILE}"

if [[ ${PURPOSE} == "local" ]]; then
  echo "Changing container registry (${SERVERLESS})"
  yq --inplace '.global.images[] |= sub("europe-docker.pkg.dev/kyma-project/", "k3d-kyma-registry.localhost:5000/")' "${VALUES_FILE}"
fi
IMAGES_SELECTOR=".global.images[] | select(key == \"function_*\") ${MAIN_ONLY_SELECTOR}"
# replace /dev/|/prod/ with /IMG_DIRECTORY/
yq --inplace "(${IMAGES_SELECTOR})|= sub (\"/dev/|/prod/\", \"/${IMG_DIRECTORY}/\") " "${VALUES_FILE}"
# replace the last :.* with :IMG_VERSION, sicne the URL can contain a port number
yq --inplace "(${IMAGES_SELECTOR}) |= sub(\":[^:]+$\",\":${IMG_VERSION}\")" "${VALUES_FILE}"
echo "==== Local Changes (${SERVERLESS}) ===="
yq '.global.images' "${VALUES_FILE}"
echo "==== End of Local Changes (${SERVERLESS}) ===="
done


# replace envs in operator
VALUES_FILE="${PROJECT_ROOT}/config/operator/base/deployment/deployment.yaml"

if [[ ${PURPOSE} == "local" ]]; then
  echo "Changing container registry (operator)"
  yq --inplace '(.spec.template.spec.containers[0].env[] | select(.name == "IMAGE_*") | .value) |= sub("europe-docker.pkg.dev/kyma-project/", "k3d-kyma-registry.localhost:5000/")' "${VALUES_FILE}"
fi

IMAGES_SELECTOR=".spec.template.spec.containers[0].env[] | select(.name == \"IMAGE_FUNCTION*\") | .value ${MAIN_ONLY_SELECTOR}"
# replace /dev/|/prod/ with /IMG_DIRECTORY/
yq --inplace "(${IMAGES_SELECTOR}) |= sub (\"/dev/|/prod/\", \"/${IMG_DIRECTORY}/\") " "${VALUES_FILE}"
yq --inplace "(${IMAGES_SELECTOR}) |= sub (\"/restricted-dev/|/restricted-prod/\", \"/restricted-${IMG_DIRECTORY}/\") " "${VALUES_FILE}"
# replace the last :.* with :IMG_VERSION, sicne the URL can contain a port number
yq --inplace "(${IMAGES_SELECTOR}) |= sub(\":[^:]+$\",\":${IMG_VERSION}\")" "${VALUES_FILE}"
echo "==== Local Changes (operator) ===="
yq '.spec.template.spec.containers[0].env[] | select(.name == "IMAGE_*")' "${VALUES_FILE}"
echo "==== End of Local Changes (operator) ===="

# update versions in labels

yq ".global.commonLabels.version |= \"${IMG_VERSION}\"" ${PROJECT_ROOT}/config/serverless/values.yaml 
yq --inplace ".appVersion |= \"${IMG_VERSION}\"" ${PROJECT_ROOT}/config/buildless-serverless/Chart.yaml
