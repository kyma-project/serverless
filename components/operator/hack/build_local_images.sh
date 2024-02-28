#!/bin/bash

ROOT_DIR="../../../.."

cd ${ROOT_DIR} | docker compose build

for i in $(yq '.services[].image' ${ROOT_DIR}/docker-compose.yaml); do
    k3d image import "$i" -c kyma
done
