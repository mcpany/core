#!/bin/bash
export GOMEMLIMIT=500MiB
export GOGC=5
export PATH=$PATH:$(go env GOPATH)/bin
cd server
golangci-lint run -c .golangci.yml -j 1 --timeout 20m --fix ./cmd/... ./pkg/... ./tests/... ./examples/...
