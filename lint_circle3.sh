#!/bin/bash
wget -qO- https://go.dev/dl/go1.26.1.linux-amd64.tar.gz | tar -xzC /tmp
export PATH=/tmp/go/bin:$PATH
export GOTOOLCHAIN=local
cd server
$(go env GOPATH)/bin/golangci-lint run --timeout 20m ./cmd/server/...
