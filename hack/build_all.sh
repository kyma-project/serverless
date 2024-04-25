#!/bin/bash
set -e
HACK_DIR="./hack"

echo "New tag: $IMG_VERSION"
(cd "${HACK_DIR}" && pwd && docker compose build)

for i in $(yq '.services[].image' ${HACK_DIR}/docker-compose.yaml); do
  IMAGE_NAME=$(echo "${i}" | cut -d ":" -f 1,2)
  NEW_IMAGE_NAME="${IMAGE_NAME}:${IMG_VERSION}"
  docker tag "${i}" "${NEW_IMAGE_NAME}"
  set -x
  docker push "${NEW_IMAGE_NAME}"
  set +x
done
