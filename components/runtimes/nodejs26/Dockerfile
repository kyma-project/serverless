# build with --build-arg BASE=non-fips or --build-arg BASE=fips to select the base image
ARG BASE=non-fips

# Node25 runtime will be updated to Node 26 as soon as it's released, this is a placeholder for the API rework
# TODO(Hx2): replace with synced nodejs26 image after it's released, 25 is the last non-LTS version
FROM node:25.9.0-alpine3.23 AS non-fips

FROM europe-docker.pkg.dev/kyma-project/restricted-prod/sap.com/node-fips:25.9.0-dev AS fips

FROM ${BASE}

ARG NODE_ENV
ENV NODE_ENV=$NODE_ENV
ENV npm_config_cache=/tmp/

USER root

RUN mkdir -p /usr/src/app
RUN mkdir -p /usr/src/app/lib
WORKDIR /usr/src/app

COPY --chown=root:root ./package.json /usr/src/app/
RUN chmod 644 /usr/src/app/package.json

RUN npm install && npm cache clean --force
COPY --chown=root:root . /usr/src/app/
RUN chmod -R 755 /usr/src/app/lib
RUN chmod 644 /usr/src/app/server.mjs

USER 1000

CMD ["npm", "start"]

EXPOSE 8888
