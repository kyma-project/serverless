FROM europe-docker.pkg.dev/kyma-project/prod/external/library/alpine:3.23.3

# https://pkgs.alpinelinux.org/packages?name=nodejs&branch=v3.20&repo=main&arch=&maintainer=
ENV NODE_VERSION=20.15.1-r0
# use Ada version compatible with the Node version
ENV ADA_VERSION=2.7.8-r0
# use NPM version compatible with the Node version
ENV NPM_VERSION=10.9.1-r0

# Update repositories to prioritize v3.20 and allow downgrades
RUN echo "http://dl-cdn.alpinelinux.org/alpine/v3.20/main" > /etc/apk/repositories && \
    echo "http://dl-cdn.alpinelinux.org/alpine/v3.20/community" >> /etc/apk/repositories && \
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
COPY --chown=root:root ./ /usr/src/app/
RUN chmod -R 755 /usr/src/app/lib
RUN chmod 644 /usr/src/app/server.mjs

CMD ["npm", "start"]

EXPOSE 8888
