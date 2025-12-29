FROM europe-docker.pkg.dev/kyma-project/prod/external/ghcr.io/nginx/alpine-fips:0.5.0-alpine3.23

# https://pkgs.alpinelinux.org/packages?name=nodejs&branch=v3.20&repo=main&arch=&maintainer=
ENV NODE_VERSION=20.17.0-r0
# use Ada version compatible with the Alpine 3.23 and Node version
ENV ADA_VERSION=3.3.0-r0

RUN apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/v3.23/main ada-libs=${ADA_VERSION} nodejs=${NODE_VERSION} npm

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
COPY --chown=root:root ./ /usr/src/app/
RUN chmod -R 755 /usr/src/app/lib
RUN chmod 644 /usr/src/app/server.mjs

CMD ["npm", "start"]

EXPOSE 8888
