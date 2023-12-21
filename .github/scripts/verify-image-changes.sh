#!/bin/bash

MAIN_IMAGES=(${MAIN_IMAGES?"Define MAIN_IMAGES env"})
PR_NOT_MAIN_IMAGES=(${PR_NOT_MAIN_IMAGES?"Define PR_NOT_MAIN_IMAGES env"})

FAIL=false
for main_image in "${MAIN_IMAGES[@]}"; do
    echo "${main_image} checking..."

    for pr_image in "${PR_NOT_MAIN_IMAGES[@]}"; do
        if [ "${main_image}" == "${pr_image}" ]; then
            echo "  warning: ${pr_image} tag/version seems to be modified (should be main)!"
            FAIL=true
        fi
    done
done

if $FAIL; then
    exit 1
fi
