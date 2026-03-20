#!/bin/bash
export PATH=/tmp/go1.26/bin:$PATH
export GOTOOLCHAIN=auto
cd server
rm -rf vendor
go mod tidy
go mod vendor
$(go env GOPATH)/bin/golangci-lint run ./cmd/server/...
