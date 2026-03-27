#!/bin/bash

set -e
set -o pipefail

export scenario=${SCENARIO:-hello_world}
export resource_namespace=${NAMESPACE:-default}
export RAPID=${RAPID:-false}

export start_timestamp="$(date +'%T')"
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
    echo -e "\nCreating function ${resource_name}"
    cat templates/function.yaml | envsubst | kubectl apply -f -

    echo -e "\nWaiting for function ${resource_name} to be ready"
    kubectl_wait function --for='jsonpath={.status.conditions[?(@.type=="Running")].status}=True' ${resource_name} -n ${resource_namespace} --timeout=2m

    # Create k6 CM
    echo -e "\nCreating k6 ConfigMap"
    cat templates/cm.yaml | envsubst | kubectl apply -f -

    # Create k6 Job
    echo -e "\nCreating k6 Job"
    cat templates/k6_job.yaml | envsubst | kubectl apply -f -

    echo -e "\nWaiting for k6 Job to complete"
    kubectl_wait job --for=condition=complete k6-${resource_name} -n ${resource_namespace} --timeout=10m

    if [ $RAPID == "false" ]; then
        # Removing resources
        echo -e "\nCleaning up resources"
        cat templates/function.yaml | envsubst | kubectl delete -f -
        cat templates/cm.yaml | envsubst | kubectl delete -f -
        cat templates/k6_job.yaml | envsubst | kubectl delete -f -
    fi
}

for variant_args in "${VARIANTS[@]}"; do
    echo "## Testing runtime: ${variant_args}"
    test_runtime $variant_args
    echo ""
done
