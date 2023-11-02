# Build the operator binary
FROM golang:1.21.3 as builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY components/operator/go.mod components/operator/go.sum ./

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY components/operator .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o operator ./main.go

# Use distroless as minimal base image to package the operator binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --chown=65532:65532 --from=builder /workspace/operator .
COPY --chown=65532:65532 config/serverless /module-chart
USER 65532:65532

ENTRYPOINT ["/operator"]