#!/bin/sh

SCRIPTDIR=$(dirname -- $0)
MODULECHART=${SCRIPTDIR}/../module-chart

${SCRIPTDIR}/clone_dir_from_github.sh kyma resources/serverless ${MODULECHART}
