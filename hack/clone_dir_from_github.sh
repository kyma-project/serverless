#!/bin/sh

## args:
# REPONAME - repo name from the kyma-project org
# REPODIR - dir inside the repo we want to clone
# TARGETDIR - full path to the dir we want to fulfill with items from REPODIR
REPONAME=$1
REPODIR=$2
TARGETDIR=$3

clone () {
  mkdir -p ${TARGETDIR}
  git clone \
    --depth 1  \
    --filter=blob:none  \
    --no-checkout \
    https://github.com/kyma-project/kyma ${TARGETDIR}
  git -C ${TARGETDIR} sparse-checkout set ${REPODIR}
  git -C ${TARGETDIR} checkout bf3e338c1609ad30d19b312ac6afe6327635e175
  
  rm ${TARGETDIR}/* \
    ${TARGETDIR}/.gitignore \
    ${TARGETDIR}/.dockerignore \
    ${TARGETDIR}/.hadolint.yaml
  mv ${TARGETDIR}/${REPODIR}/* ${TARGETDIR}
  rm -rf ${TARGETDIR}/$(echo "${REPODIR}" | cut -d "/" -f1)

  echo $(basename ${TARGETDIR}) generated
}

upgrade () {
  LASTREMOTECOMMIT=$(curl -Ss "https://api.github.com/repos/kyma-project/${REPONAME}/commits/main" | jq -r '.sha')
  LASTLOCALCOMMIT=$(git -C ${TARGETDIR} log --pretty=format:"%H")
  if [ "${LASTREMOTECOMMIT}" != "${LASTLOCALCOMMIT}" ];
  then
    echo "updating $(basename ${TARGETDIR}) to commit ${LASTREMOTECOMMIT}"
    rm -rf ${TARGETDIR}
    clone
  else
    echo "$(basename ${TARGETDIR}) is up-to-date"
  fi
}

# check if dir doesn't exist or is empty
if [ ! -d "${TARGETDIR}" ] || [ ! "$(ls -A ${TARGETDIR})" ];
then
  clone
else
  upgrade
fi
