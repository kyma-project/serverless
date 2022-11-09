#!/bin/sh

kyma provision k3d
kyma deploy -s main -p production \
    --component cluster-essentials \
    --component istio-resources \
    --component istio
