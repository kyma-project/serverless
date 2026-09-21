#! /bin/sh
set -e;

echo "{}" > package.json;

# copy the code, either from the mounted git repo or provided inline source
if [ -d "/git-repository" ]; then
  cp -r /git-repository/src/* .;
else 
  echo "${FUNC_HANDLER_SOURCE}" > handler.js
  if [ -n "${FUNC_HANDLER_DEPENDENCIES}" ]; then
    echo "${FUNC_HANDLER_DEPENDENCIES}" > package.json
  fi
fi

# install packages
NPM_CONFIG_USERCONFIG=package-registry-config/.npmrc npm install --prefer-offline --no-audit --progress=false;

# start the server
cd ..;
npm start;
