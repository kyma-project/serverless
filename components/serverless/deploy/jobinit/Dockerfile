FROM europe-docker.pkg.dev/kyma-project/prod/external/library/golang:1.25.7-alpine3.23 AS builder

ENV BASE_APP_DIR=/workspace/go/src/github.com/kyma-project/serverless \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64 \
    GOFIPS140=v1.0.0 \
    LIBGIT2_VERSION=1.5.2-r0

RUN apk add --no-cache gcc libc-dev
RUN apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/v3.18/community libgit2-dev=${LIBGIT2_VERSION}

WORKDIR ${BASE_APP_DIR}

#
# copy files allowed in .dockerignore
#
COPY . ${BASE_APP_DIR}/

RUN go build -ldflags "-s -w" -a -o jobinit ./components/serverless/cmd/jobinit/main.go \
    && mkdir /app \
    && mv ./jobinit /app/jobinit

FROM europe-docker.pkg.dev/kyma-project/prod/external/library/alpine:3.23.3
ENV LIBGIT2_VERSION=1.5.2-r0 \
    GODEBUG=fips140=on,tlsmlkem=0

LABEL source=git@github.com:kyma-project/serverless.git

RUN apk update --no-cache && apk upgrade --no-cache
RUN apk add --no-cache ca-certificates
RUN apk add --no-cache --update --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main openssh-client openssl
RUN apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/v3.18/community libgit2=${LIBGIT2_VERSION}

COPY --from=builder /app /app

ENTRYPOINT ["/app/jobinit"]
