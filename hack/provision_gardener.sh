#!/bin/bash

set -eo pipefail

# render and applyshoot template
shoot_template=$(envsubst < ${PROJECT_ROOT}/hack/shoot_template.yaml)

export KUBECONFIG="${GARDENER_SA_PATH}"
echo "$shoot_template" | kubectl --kubeconfig=$KUBECONFIG apply -f -

echo "waiting fo cluster to be ready..."
kubectl --kubeconfig=$KUBECONFIG wait --for=condition=EveryNodeReady shoot/${SHOOT} --timeout=17m

# create kubeconfig request, that creates a kubeconfig which is valid for one day
kubectl --kubeconfig=$KUBECONFIG create \
    -f <(printf '{"spec":{"expirationSeconds":86400}}') \
    --raw /apis/core.gardener.cloud/v1beta1/namespaces/garden-${PROJECT}/shoots/${SHOOT}/adminkubeconfig | \
    jq -r ".status.kubeconfig" | \
    base64 -d > ${SHOOT}_kubeconfig.yaml

# replace the default kubeconfig
mkdir -p ~/.kube
mv ${SHOOT}_kubeconfig.yaml ~/.kube/config
