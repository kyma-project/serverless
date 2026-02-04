FROM europe-docker.pkg.dev/kyma-project/prod/external/library/alpine:3.23.3

# https://pkgs.alpinelinux.org/packages?name=nodejs&branch=v3.23&repo=main&arch=&maintainer=
ENV NODE_VERSION=24.13.0-r1
# use Ada version compatible with the Node version
ENV ADA_VERSION=3.3.0-r0
# use NPM version compatible with the Node version
ENV NPM_VERSION=11.6.3-r0

# Update repositories to prioritize v3.23 and allow downgrades
RUN echo "http://dl-cdn.alpinelinux.org/alpine/v3.23/main" > /etc/apk/repositories && \
    echo "http://dl-cdn.alpinelinux.org/alpine/v3.23/community" >> /etc/apk/repositories && \
    apk update && \
    apk add --no-cache --allow-untrusted --force-overwrite ada-libs=${ADA_VERSION} nodejs=${NODE_VERSION} npm=${NPM_VERSION}

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
