#!/usr/bin/env bash

collect_logs(){
    set +o errexit
    echo "####################"
    echo "kubectl get pods -A"
    echo "###################"
    kubectl get pods -A

    echo "########################"
    echo "kubectl get functions -A"
    echo "########################"
    kubectl get functions -A

    echo "########################################################"
    echo "kubectl logs -n kyma-system -l app=serverless --tail=-1"
    echo "########################################################"
    kubectl logs -n kyma-system -l app=serverless --tail=-1


    echo "########################################################"
    echo "kubectl logs -n kyma-system -l app=serverless-webhook --tail=-1"
    echo "########################################################"
    kubectl logs -n kyma-system -l app=serverless-webhook --tail=-1


    echo "########################################################"
    echo "Get logs from all function build jobs and runtime"
    echo "########################################################"
    ALL_TEST_NAMESPACES=$(kubectl get namespace --selector created-by=serverless-controller-manager-test   --no-headers -o custom-columns=name:.metadata.name)
    # shellcheck disable=SC2206
    ALL=($ALL_TEST_NAMESPACES)
    for NAMESPACE in "${ALL[@]}"
    do
      kubectl logs --namespace "${NAMESPACE}" --all-containers  --selector serverless.kyma-project.io/function-name --ignore-errors --prefix=true
    done
    echo ""
    set -o errexit
}
