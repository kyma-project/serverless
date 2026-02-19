#!/bin/bash

IMG_DIRECTORY=${IMG_DIRECTORY?"Define IMG_DIRECTORY env"}
IMG_VERSION=${IMG_VERSION?"Define IMG_VERSION env"}
PROJECT_ROOT=${PROJECT_ROOT?"Define PROJECT_ROOT env"}

CONFIG_DEV_DIR=${PROJECT_ROOT}/config/operator/dev
CONFIG_BASE_DIR=${PROJECT_ROOT}/config/operator/base
DEFAULT_IMAGES_PATCH=${CONFIG_DEV_DIR}/default-images-patch.yaml
OPERATOR_DEPLOYMENT=${CONFIG_BASE_DIR}/deployment/deployment.yaml

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
