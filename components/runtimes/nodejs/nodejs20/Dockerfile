# image base on node:20.13.1-alpine3.20
FROM node@sha256:b3c9885d73a67347e607a24ce5487287caea8261cb7e0f0cadf12c0aedd7874c

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

COPY --chown=root:root ./server.mjs /usr/src/app/server.mjs
RUN chmod 644 /usr/src/app/server.mjs

CMD ["npm", "start"]

EXPOSE 8888
