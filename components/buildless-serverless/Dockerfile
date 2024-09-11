FROM europe-docker.pkg.dev/kyma-project/prod/external/library/golang:1.22.6-alpine as builder

ENV BASE_APP_DIR=/workspace/go/src/github.com/kyma-project/buildless-serverless

WORKDIR ${BASE_APP_DIR}

#
# copy files allowed in .dockerignore
#
COPY . ${BASE_APP_DIR}/

RUN go build -o manager ./components/buildless-serverless/cmd/main.go \
    && mkdir /app \
    && mv ./manager /app/manager

FROM europe-docker.pkg.dev/kyma-project/prod/external/library/alpine:3.20.2 as certs

RUN apk add --no-cache ca-certificates

FROM europe-docker.pkg.dev/kyma-project/prod/external/library/alpine:3.20.2

LABEL source = git@github.com:kyma-project/serverless.git

COPY --from=builder /app /app
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["/app/manager"]

