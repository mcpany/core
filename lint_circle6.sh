#!/bin/bash
wget -qO- https://go.dev/dl/go1.26.1.linux-amd64.tar.gz | tar -xzC /tmp
export PATH=/tmp/go/bin:$PATH
export GOTOOLCHAIN=go1.26.1
cd server
git checkout origin/main -- go.mod go.work .golangci.yml
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.6
/tmp/go/bin/golangci-lint run --timeout 20m ./cmd/server/...
