FROM cgr.dev/sap.com/node-fips:22-dev

USER root

RUN mkdir -p /usr/src/app
RUN mkdir -p /usr/src/app/lib
WORKDIR /usr/src/app

COPY --chown=root:root ./package.json /usr/src/app/
RUN chmod 644 /usr/src/app/package.json

RUN npm install && npm cache clean --force
COPY --chown=root:root . /usr/src/app/
RUN chmod -R 755 /usr/src/app/lib
RUN chmod 644 /usr/src/app/server.mjs

USER 1000

CMD ["npm", "start"]

EXPOSE 8888