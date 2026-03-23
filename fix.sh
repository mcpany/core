#!/bin/bash
git fetch origin main
git checkout main
export GOPATH=/tmp/go
export GOCACHE=/tmp/gocache
export GOLANGCI_LINT_CACHE=/tmp/golangci-lint
export GOMAXPROCS=1
cd server
/app/build/env/bin/golangci-lint run --timeout 30m --concurrency 1 ./pkg/...
