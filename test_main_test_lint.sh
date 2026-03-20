#!/bin/bash
cd server
rm -rf /tmp/go1.26
mkdir -p /tmp/go1.26
wget -qO- https://go.dev/dl/go1.24.1.linux-amd64.tar.gz | tar -xzC /tmp/go1.26 --strip-components=1
export PATH=/tmp/go1.26/bin:$PATH
export GOTOOLCHAIN=local
go install github.com/mgechev/revive@latest
$(go env GOPATH)/bin/revive -config .golangci-revive.toml -formatter friendly ./cmd/server/... || $(go env GOPATH)/bin/revive -formatter friendly ./cmd/server/...
