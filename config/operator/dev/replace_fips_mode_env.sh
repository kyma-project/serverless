#!/bin/bash

# This script replaces images in default-images-patch.yaml with the ones from operator deployment and then with the ones provided as input parameters.
# It can work in two modes - when IMG_DIRECTORY and IMG_VERSION are:
# - not set: do nothing, images from operator's chart will be used
# - set: replace images in default-images-patch.yaml, images from operator's envs will be used


PROJECT_ROOT=${PROJECT_ROOT?"Define PROJECT_ROOT env"}

FIPS_MODE_PATCH=${PROJECT_ROOT}/config/operator/dev/fips-mode-patch.yaml

echo "Enabling fips mode in ${FIPS_MODE_PATCH}"

# replace images in images patch with current images from operator
yq --inplace '.spec.template.spec.containers[0].env[0].value = "true"' ${FIPS_MODE_PATCH}

