FROM europe-docker.pkg.dev/kyma-project/prod/external/library/alpine:3.21.3

# https://pkgs.alpinelinux.org/packages?name=nodejs&branch=v3.21&repo=main&arch=&maintainer=
ENV NODE_VERSION=22.13.1-r0

RUN apk add --no-cache openssl3 nodejs=${NODE_VERSION} npm

ARG NODE_ENV
ENV NODE_ENV=$NODE_ENV
ENV npm_config_cache=/tmp/

RUN mkdir -p /usr/src/app
RUN mkdir -p /usr/src/app/lib
WORKDIR /usr/src/app

COPY --chown=root:root ./nodejs22/package.json /usr/src/app/
RUN chmod 644 /usr/src/app/package.json

RUN npm install && npm cache clean --force
COPY --chown=root:root ./lib /usr/src/app/lib
RUN chmod -R 755 /usr/src/app/lib

COPY --chown=root:root ./server.mjs /usr/src/app/server.mjs
RUN chmod 644 /usr/src/app/server.mjs

CMD ["npm", "start"]

EXPOSE 8888
