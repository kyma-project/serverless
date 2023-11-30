#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
ROOT_DIR="${SCRIPT_DIR}/../../.."

cd $ROOT_DIR | docker compose build

for i in $(yq '.services[].image' $ROOT_DIR/docker-compose.yml); do
    k3d image import "$i" -c kyma
done

${SCRIPT_DIR}/poc_script_replace_images.sh "dev" "local"
