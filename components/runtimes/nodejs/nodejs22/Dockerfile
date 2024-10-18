FROM europe-docker.pkg.dev/kyma-project/prod/external/library/alpine:3.20.3

ENV NODE_VERSION=22.10.0


RUN apk add --no-cache openssl3 libgcc libstdc++
RUN wget -O node.tar.xz https://unofficial-builds.nodejs.org/download/release/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-x64-musl.tar.xz && \
  tar -xJf node.tar.xz && \
  cp -RT node-v${NODE_VERSION}-linux-x64-musl/bin /usr/bin && \
  cp -RT node-v${NODE_VERSION}-linux-x64-musl/include /usr/include && \
  cp -RT node-v${NODE_VERSION}-linux-x64-musl/lib /usr/lib && \
  cp -RT node-v${NODE_VERSION}-linux-x64-musl/share /usr/share && \
  rm -rf node-v${NODE_VERSION}-linux-x64-musl && \
  rm node.tar.xz 

ARG NODE_ENV
ENV NODE_ENV $NODE_ENV
ENV npm_config_cache /tmp/

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
