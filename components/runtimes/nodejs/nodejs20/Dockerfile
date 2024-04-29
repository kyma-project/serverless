# image base on node:20.12.1-alpine3.19
FROM node@sha256:7e227295e96f5b00aa79555ae166f50610940d888fc2e321cf36304cbd17d7d6

ARG NODE_ENV
ENV NODE_ENV $NODE_ENV
ENV npm_config_cache /tmp/

RUN mkdir -p /usr/src/app
RUN mkdir -p /usr/src/app/lib
WORKDIR /usr/src/app

COPY --chown=root:root ./nodejs20/package.json /usr/src/app/
RUN chmod 644 /usr/src/app/package.json

RUN npm install && npm cache clean --force
COPY --chown=root:root ./lib /usr/src/app/lib
RUN chmod -R 755 /usr/src/app/lib

COPY --chown=root:root ./server.js /usr/src/app/server.js
RUN chmod 644 /usr/src/app/server.js

CMD ["npm", "start"]

EXPOSE 8888
