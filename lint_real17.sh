#!/bin/bash
export PATH=/tmp/go1.26/bin:$PATH
export GOTOOLCHAIN=local
cd server
$(go env GOPATH)/bin/golangci-lint run --timeout=10m ./cmd/server/...
