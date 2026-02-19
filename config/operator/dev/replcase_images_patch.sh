#!/bin/bash

# This script replaces images in default-images-patch.yaml with the ones from operator deployment and then with the ones provided as input parameters.
# It can work in two modes - when IMG_DIRECTORY and IMG_VERSION are:
# - not set: do nothing, images from operator's chart will be used
# - set: replace images in default-images-patch.yaml, images from operator's envs will be used

IMG_DIRECTORY=${IMG_DIRECTORY:-""}
IMG_VERSION=${IMG_VERSION:-""}
PROJECT_ROOT=${PROJECT_ROOT?"Define PROJECT_ROOT env"}

if [ "${IMG_DIRECTORY}" = "" ] || [ "${IMG_VERSION}" = "" ]; then
    echo "Input parameters are not set - skipping images replacement. Images from chart will be used"
    exit 0
fi

OPERATOR_DEPLOYMENT=${PROJECT_ROOT}/config/operator/base/deployment/deployment.yaml
DEFAULT_IMAGES_PATCH=${PROJECT_ROOT}/config/operator/dev/default-images-patch.yaml

echo "Replacing images in ${DEFAULT_IMAGES_PATCH} with directory ${IMG_DIRECTORY} and version ${IMG_VERSION}"

# get current images from operatror as template for replacement
DEPLOY_ENVS="$(yq '.spec.template.spec.containers[0].env | filter(.name == "IMAGE_FUNCTION_*")' ${OPERATOR_DEPLOYMENT})"

# replace images in images patch with current images from operator
yq --inplace '.spec.template.spec.containers[0].env |= env(DEPLOY_ENVS)' ${DEFAULT_IMAGES_PATCH}

# replace images in images patch with desired ones
IMAGES_SELECTOR=".spec.template.spec.containers[0].env[] | select(.name == \"IMAGE_FUNCTION*\") | .value ${MAIN_ONLY_SELECTOR}"
# replace /dev/|/prod/ with /IMG_DIRECTORY/ and the same for restricted ones
yq --inplace "(${IMAGES_SELECTOR}) |= sub (\"/dev/|/prod/\", \"/${IMG_DIRECTORY}/\") " "${DEFAULT_IMAGES_PATCH}"
yq --inplace "(${IMAGES_SELECTOR}) |= sub (\"/restricted-dev/|/restricted-prod/\", \"/restricted-${IMG_DIRECTORY}/\") " "${DEFAULT_IMAGES_PATCH}"
# replace the last :.* with :IMG_VERSION, sicne the URL can contain a port number
yq --inplace "(${IMAGES_SELECTOR}) |= sub(\":[^:]+$\",\":${IMG_VERSION}\")" "${DEFAULT_IMAGES_PATCH}"
