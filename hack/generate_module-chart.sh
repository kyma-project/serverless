#!/bin/sh

SCRIPTDIR=$(dirname -- $0)
MODULECHART=${SCRIPTDIR}/../module-chart

mkdir -p ${MODULECHART}
git clone \
  --depth 1  \
  --filter=blob:none  \
  --no-checkout \
  https://github.com/kyma-project/kyma ${MODULECHART}
git -C ${MODULECHART} sparse-checkout set resources/serverless
git -C ${MODULECHART} checkout main
mv ${MODULECHART}/resources/serverless/* ${MODULECHART}
rm -rf ${MODULECHART}/.git
rm -rf ${MODULECHART}/resources
echo module-chart generated
