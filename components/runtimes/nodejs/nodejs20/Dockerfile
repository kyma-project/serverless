# image base on node:20.12.1-alpine3.19
FROM node@sha256:7e227295e96f5b00aa79555ae166f50610940d888fc2e321cf36304cbd17d7d6

USER root

ARG NODE_ENV
ENV NODE_ENV $NODE_ENV
ENV npm_config_cache /tmp/

RUN mkdir -p /usr/src/app
RUN mkdir -p /usr/src/app/lib
WORKDIR /usr/src/app

COPY ./nodejs20/package.json /usr/src/app/
RUN npm install && npm cache clean --force
COPY ./lib /usr/src/app/lib

COPY ./server.js /usr/src/app/server.js

CMD ["npm", "start"]

EXPOSE 8888
