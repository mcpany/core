# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

# Build the manager binary
FROM golang:1.24 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
COPY operator/go.mod operator/go.mod
COPY operator/go.sum operator/go.sum

# cache deps
WORKDIR /workspace/operator
RUN go mod download

# Copy the go source
WORKDIR /workspace
COPY server server
COPY operator operator

# Build
WORKDIR /workspace/operator
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager cmd/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/operator/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
