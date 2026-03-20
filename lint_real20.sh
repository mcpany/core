#!/bin/bash
export PATH=/tmp/go1.26/bin:$PATH
export GOTOOLCHAIN=auto
export GOWORK=off
cd server
make test-proto
$(go env GOPATH)/bin/golangci-lint run ./cmd/server/...
