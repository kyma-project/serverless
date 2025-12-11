#!/bin/bash

#
## INPUTS
#
# replace these with your local paths
KEDA_REPO=/Users/I517616/go/src/github.com/kyma-project/keda-manager
SERVERLESS_REPO=/Users/I517616/go/src/github.com/kyma-project/serverless-manager

#
## IMAGES
#

KEDA=cgr.dev/sap.com/keda-fips:2.18.2
KEDA_METRICS_APISERVER=cgr.dev/sap.com/keda-metrics-apiserver-fips:v2.18.2
KEDA_WEBHOOK=cgr.dev/sap.com/keda-admission-webhooks-fips:v2.18.2
NODE=cgr.dev/sap.com/node-fips:22-dev
PYTHON=cgr.dev/sap.com/python-fips:3.13-dev

#
## SETUP CLUSTER
#

k3d cluster create \
    --kubeconfig-switch-context \
    --port 80:80@loadbalancer \
    --port 443:443@loadbalancer \
    --image rancher/k3s:v1.33.5-k3s1 \
    --k3s-arg "--disable=traefik@server:*"
kubectl create namespace kyma-system
kubectl label namespace kyma-system istio-injection=enabled --overwrite

#
## ISTIO
#

kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-manager.yaml
kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-default-cr.yaml
kubectl wait --for='jsonpath={.status.state}=Ready' istio/default -n kyma-system

#
## KEDA
#

# import keda operator
docker pull ${KEDA}
k3d image import ${KEDA}
yq -i ".spec.template.spec.containers[0].env[] |= select(.name==\"IMAGE_KEDA_OPERATOR\") |= .value=\"${KEDA}\"" \
    ${KEDA_REPO}/config/manager/manager.yaml

# import keda metrics apiserver
docker pull ${KEDA_METRICS_APISERVER}
k3d image import ${KEDA_METRICS_APISERVER}
yq -i ".spec.template.spec.containers[0].env[] |= select(.name==\"IMAGE_KEDA_METRICS_APISERVER\") |= .value=\"${KEDA_METRICS_APISERVER}\"" \
    ${KEDA_REPO}/config/manager/manager.yaml

# import keda admission webhook
docker pull ${KEDA_WEBHOOK}
k3d image import ${KEDA_WEBHOOK}
yq -i ".spec.template.spec.containers[0].env[] |= select(.name==\"IMAGE_KEDA_ADMISSION_WEBHOOKS\") |= .value=\"${KEDA_WEBHOOK}\"" \
    ${KEDA_REPO}/config/manager/manager.yaml

# setup required envs for all keda deployments
yq -i ea '. |= [.].[] | select(.kind=="Deployment") |= .spec.template.spec.containers[0].env += {"name": "GODEBUG", "value": "fips140=v1.0.0,tlsmlkem=0"}' \
    ${KEDA_REPO}/keda.yaml

# keda installation
IMG=europe-docker.pkg.dev/kyma-project/prod/keda-manager:main make -C ${KEDA_REPO} install deploy
kubectl apply -f ${KEDA_REPO}/config/samples/keda-default-cr.yaml -n kyma-system

#
## SERVERLESS
#

# Python312
PYTHON_DOCKERFILE=$(cat ${SERVERLESS_REPO}/components/runtimes/python312/Dockerfile | sed "1s|.*|FROM ${PYTHON}|")
printf "$PYTHON_DOCKERFILE" > ${SERVERLESS_REPO}/components/runtimes/python312/Dockerfile
docker pull ${PYTHON}
docker build -t python-fips:1 ${SERVERLESS_REPO}/components/runtimes/python312
k3d image import python-fips:1
yq -i ". |= .spec.template.spec.containers[0].env[] |= select(.name==\"IMAGE_FUNCTION_RUNTIME_PYTHON312\") |= .value=\"python-fips:1\"" ${SERVERLESS_REPO}/config/operator/dev/default-images-patch.yaml

# Nodejs22
NODE_DOCKERFILE=$(cat ${SERVERLESS_REPO}/components/runtimes/nodejs22/Dockerfile | sed "1s|.*|FROM ${NODE}|")
printf "$NODE_DOCKERFILE" > ${SERVERLESS_REPO}/components/runtimes/nodejs22/Dockerfile
docker pull ${NODE}
docker build -t node-fips:1 ${SERVERLESS_REPO}/components/runtimes/nodejs22
k3d image import node-fips:1
yq -i ". |= .spec.template.spec.containers[0].env[] |= select(.name==\"IMAGE_FUNCTION_RUNTIME_NODEJS22\") |= .value=\"node-fips:1\"" ${SERVERLESS_REPO}/config/operator/dev/default-images-patch.yaml

# Controller
docker build -t controller:1 -f ${SERVERLESS_REPO}/components/buildless-serverless/controller ${SERVERLESS_REPO}
yq -i ". |= .spec.template.spec.containers[0].env[] |= select(.name==\"IMAGE_FUNCTION_BUILDLESS_CONTROLLER\") |= .value=\"controller:1\"" ${SERVERLESS_REPO}/config/operator/dev/default-images-patch.yaml
k3d image import controller:1

# install serverless
IMG=europe-docker.pkg.dev/kyma-project/prod/serverless-operator:main make -C ${SERVERLESS_REPO} install-serverless-custom-operator
