#!/bin/bash

set -e
set -o pipefail

export scenario=${SCENARIO:-hello_world}
export resource_namespace=${NAMESPACE:-default}
export RAPID=${RAPID:-false}

export start_date="$(date +'%Y-%m-%d')"
export start_timestamp="$(date +'%T')"
export results_cm="k6-results-${start_date}-$(echo ${start_timestamp} | tr ':' '-')"
export results_cm_ns="default"
export result_dir="/tmp/${results_cm}"
VARIANTS=(
    "nodejs22 XS"
    "nodejs22 S"
    "nodejs22 M"
    "nodejs22 L"
    "nodejs22 XL"
    "nodejs24 XS"
    "nodejs24 S"
    "nodejs24 M"
    "nodejs24 L"
    "nodejs24 XL"
    "python312 XS"
    "python312 S"
    "python312 M"
    "python312 L"
    "python312 XL"
)

kubectl_wait() {
    if [ $RAPID == "true" ]; then
        echo "RAPID mode enabled, skipping wait"
    else
        kubectl wait "$@"
    fi
}

kubectl_logs() {
    if [ $RAPID == "true" ]; then
        echo "RAPID mode enabled, skipping logs"
    else
        kubectl logs "$@"
    fi
}

test_runtime(){
    export runtime=${1:?Runtime name is required (e.g. nodejs24, python312)}
    export runtime_profile=${2:?Runtime profile is required (e.g. XS, S, M, L, XL)}
    export testid="${runtime}-$( echo $runtime_profile | tr '[:upper:]' '[:lower:]')"
    export resource_name="${testid}"
    code_dir="scenarios/${scenario}"

    export k6_source="$(cat $code_dir/k6.js | sed 's/^/        /')"

    if [[ $runtime == python* ]]; then
        export runtime_family="python"
        export dependencies="$(cat $code_dir/requirements.txt | sed 's/^/        /')"
        export source="$(cat $code_dir/handler.py | sed 's/^/        /')"
    else
        export runtime_family="nodejs"
        export dependencies="$(cat $code_dir/package.json | sed 's/^/        /')"
        export source="$(cat $code_dir/handler.js | sed 's/^/        /')"
    fi

    # Create function
    echo -e "\n[SCRIPT] Creating function ${resource_name}"
    cat templates/function.yaml | envsubst | kubectl apply -f -

    echo -e "\n[SCRIPT] Waiting for function ${resource_name} to be ready"
    kubectl_wait function --for='jsonpath={.status.conditions[?(@.type=="Running")].status}=True' ${resource_name} -n ${resource_namespace} --timeout=2m

    # Create k6 CM
    echo -e "\n[SCRIPT] Creating k6 ConfigMap"
    cat templates/cm.yaml | envsubst | kubectl apply -f -

    # Create k6 Job
    echo -e "\n[SCRIPT] Creating k6 Job"
    cat templates/k6_job.yaml | envsubst | kubectl apply -f -

    echo -e "\n[SCRIPT] Waiting for k6 Job to complete"
    kubectl_wait job --for=condition=complete k6-${resource_name} -n ${resource_namespace} --timeout=10m

    echo -e "\n[SCRIPT] Saving k6 results to configmap ${results_cm}"
    result_file="${result_dir}/${testid}"
    kubectl logs job/k6-${resource_name} -n ${resource_namespace} --tail=-1 > ${result_file}
    kubectl create configmap ${results_cm} -n ${results_cm_ns} --from-file=${result_dir} --dry-run=client -o yaml | kubectl apply --server-side -f -

    if [ $RAPID == "false" ]; then
        # Removing resources
        echo -e "\n[SCRIPT] Cleaning up resources"
        cat templates/function.yaml | envsubst | kubectl delete -f -
        cat templates/cm.yaml | envsubst | kubectl delete -f -
        cat templates/k6_job.yaml | envsubst | kubectl delete -f -
    fi
}

mkdir -p ${result_dir}

for variant_args in "${VARIANTS[@]}"; do
    echo -e "\n[SCRIPT] Testing runtime: ${variant_args}"
    test_runtime $variant_args
done
