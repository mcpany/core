#!/bin/bash
wget -qO- https://go.dev/dl/go1.26.1.linux-amd64.tar.gz | tar -xzC /tmp
export PATH=/tmp/go/bin:$PATH
export GOTOOLCHAIN=local
cd server
wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /tmp/go/bin v1.64.5
/tmp/go/bin/golangci-lint run --timeout 20m ./cmd/server/...
