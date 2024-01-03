#!/bin/bash

kyma provision k3d

kubectl create namespace kyma-system
operatorManifest=$(curl -L 'https://github.com/kyma-project/serverless/releases/download/1.2.1/serverless-operator.yaml')

# install serverless operator
echo "$operatorManifest" | kubectl apply -f -
kubectl scale --replicas=0 -n kyma-system deploy/serverless-operator

# create 2k secrets with huge data
x=1
secretData="${operatorManifest}${operatorManifest}${operatorManifest}${operatorManifest}${operatorManifest}${operatorManifest}${operatorManifest}${operatorManifest}${operatorManifest}${operatorManifest}${operatorManifest}${operatorManifest}"
while [ $x -le 2000 ]; do
    kubectl create secret generic secret-$x --from-literal data="${secretData}"
    x=$(( $x + 1 ))
done

# install serverless
curl -L 'https://github.com/kyma-project/serverless/releases/download/1.2.1/default-serverless-cr-k3d.yaml' | kubectl apply -f -