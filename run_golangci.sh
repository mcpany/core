#!/bin/bash
wget -qO- https://go.dev/dl/go1.26.1.linux-amd64.tar.gz | tar -xzC /tmp
export PATH=/tmp/go/bin:$PATH
export GOTOOLCHAIN=local
export GOROOT=/tmp/go
cd server
git checkout origin/main -- go.mod .golangci.yml ../go.work
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.6
/home/jules/go/bin/golangci-lint run ./cmd/server/...
