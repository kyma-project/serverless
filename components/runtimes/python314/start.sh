#! /bin/sh
set -e;

echo "" > requirements.txt;

# copy the code, either from the mounted git repo oor provided inline source
if [ -d "/git-repository" ]; then
  cp -r /git-repository/src/* .;
else 
  echo "${FUNC_HANDLER_SOURCE}" > handler.py
  echo "${FUNC_HANDLER_DEPENDENCIES}" > requirements.txt
fi

# install packages
export PYTHONPATH="/usr/src/app/.local:${PYTHONPATH}"
PIP_CONFIG_FILE=package-registry-config/pip.conf pip install --target=/usr/src/app/.local --no-cache-dir -r requirements.txt;

# start the server
cd ..;
python server.py;
