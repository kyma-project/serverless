#!/bin/sh

IMG_VERSION=${IMG_VERSION?"Define IMG_VERSION env"}

# Load both files simultaneously (fileIndex 0 = sec-scanners-config.yaml, fileIndex 1 = deployment.yaml).
# Extract all IMAGE_FUNCTION* env values from the operator deployment - these are the function runtime
# and controller images that need security scanning. Image versions are taken from deployment.yaml,
# which is updated by replace_serverless_chart_images.sh before this script runs.
# Append the serverless-operator image with the given release version.
# Replace the bdba list in sec-scanners-config.yaml with the result and discard deployment.yaml from output.
yq eval-all --inplace '
    select(fileIndex == 0).bdba=(
        [
            (
                select(fileIndex == 1)
                | .spec.template.spec.containers[0].env[]
                | select(.name | test("IMAGE_FUNCTION"))
                | .value
            ),
            "europe-docker.pkg.dev/kyma-project/prod/serverless-operator:"+ env(IMG_VERSION)
        ] | unique
    )
    | select(fileIndex == 0)
    ' sec-scanners-config.yaml config/operator/base/deployment/deployment.yaml
