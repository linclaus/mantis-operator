# Build the log-metirc-api binary
FROM golang:1.13 as builder

WORKDIR /workspace
ENV GOPROXY=https://goproxy.cn,direct
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/api/log-metric.go main.go
COPY api/ api/
COPY pkg/ pkg/
COPY controllers/ controllers/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o mantis-core-log-metirc-api main.go

# Use distroless as minimal base image to package the log-metirc-api binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM alpine:3.11
WORKDIR /
COPY --from=builder /workspace/mantis-core-log-metirc-api .

ENTRYPOINT ["/mantis-core-log-metirc-api"]
