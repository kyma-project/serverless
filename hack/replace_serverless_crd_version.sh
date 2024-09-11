#!/bin/bash

KUSTOMIZATION_DIRECTORY=${PROJECT_ROOT}/components/serverless/config/crd
KUSTOMIZATION_FILE=${KUSTOMIZATION_DIRECTORY}/kustomization.yaml

REQUIRED_ENV_VARIABLES=('IMG_VERSION' 'PROJECT_ROOT')
for VAR in "${REQUIRED_ENV_VARIABLES[@]}"; do
  if [ -z "${!VAR}" ]; then
    echo "${VAR} is undefined"
    exit 1
  fi
done

VERSION_SELECTOR='.commonLabels."app.kubernetes.io/version"'
yq --inplace "${VERSION_SELECTOR} = \"${IMG_VERSION}\"" ${KUSTOMIZATION_FILE}

make -C ${PROJECT_ROOT}/components/serverless manifests
