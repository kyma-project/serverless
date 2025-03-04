#!/usr/bin/env bash

# standard bash error handling
set -o nounset  # treat unset variables as an error and exit immediately.
set -o errexit  # exit immediately when a command fails.
set -E          # needs to be set if we want the ERR trap
set -o pipefail # prevents errors in a pipeline from being masked

# Expected variables:
GITHUB_TOKEN=${GITHUB_TOKEN?"Define GITHUB_TOKEN env"} # github token used to upload the template yaml
RELEASE_ID=${RELEASE_ID?"Define RELEASE_ID env"} # github token used to upload the template yaml

uploadFile() {
  filePath=${1}
  ghAsset=${2}

  echo "Uploading ${filePath} as ${ghAsset}"
  response=$(curl -s -o output.txt -w "%{http_code}" \
                  --request POST --data-binary @"$filePath" \
                  -H "Authorization: token $GITHUB_TOKEN" \
                  -H "Content-Type: text/yaml" \
                   $ghAsset)
  if [[ "$response" != "201" ]]; then
    echo "Unable to upload the asset ($filePath): "
    echo "HTTP Status: $response"
    cat output.txt
    exit 1
  else
    echo "$filePath uploaded"
  fi
}

make -C components/operator/ render-manifest

echo "Generated serverless-operator.yaml:"
cat serverless-operator.yaml

echo "Updating github release with assets"
UPLOAD_URL="https://uploads.github.com/repos/kyma-project/serverless/releases/${RELEASE_ID}/assets"

uploadFile "serverless-operator.yaml" "${UPLOAD_URL}?name=serverless-operator.yaml"
uploadFile "config/samples/default-serverless-cr.yaml" "${UPLOAD_URL}?name=default-serverless-cr.yaml"
uploadFile "config/samples/default-serverless-cr-k3d.yaml" "${UPLOAD_URL}?name=default-serverless-cr-k3d.yaml"



