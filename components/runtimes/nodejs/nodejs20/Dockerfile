# image base on node:20.11.1-alpine3.19 (node:20.11.1, alpine:3.19.1)
FROM node@sha256:c0a3badbd8a0a760de903e00cedbca94588e609299820557e72cba2a53dbaa2c

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
