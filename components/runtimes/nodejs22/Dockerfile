FROM europe-docker.pkg.dev/kyma-project/prod/external/ghcr.io/nginx/alpine-fips:0.4.0-alpine3.22

# https://pkgs.alpinelinux.org/packages?name=nodejs&branch=v3.21&repo=main&arch=&maintainer=
ENV NODE_VERSION=22.16.0-r2

RUN apk add --no-cache nodejs=${NODE_VERSION} npm

# FIPS
COPY openssl.cnf /etc/ssl/openssl.cnf
ENV OPENSSL_FIPS=1
ENV OPENSSL_DIR=/usr
ENV OPENSSL_CONF=/etc/ssl/openssl.cnf
ENV OPENSSL_MODULES=/usr/lib/ossl-modules

ARG NODE_ENV
ENV NODE_ENV=$NODE_ENV
ENV npm_config_cache=/tmp/

RUN mkdir -p /usr/src/app
RUN mkdir -p /usr/src/app/lib
WORKDIR /usr/src/app

COPY --chown=root:root ./package.json /usr/src/app/
RUN chmod 644 /usr/src/app/package.json

RUN npm install && npm cache clean --force
COPY --chown=root:root . /usr/src/app/
RUN chmod -R 755 /usr/src/app/lib
RUN chmod 644 /usr/src/app/server.mjs

CMD ["npm", "start"]

EXPOSE 8888
