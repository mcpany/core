#!/bin/bash
set -e
export PATH=$PATH:$(go env GOPATH)/bin
cd server

# Try to run standard staticcheck or similar as golangci-lint has issues
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./... || true

# Try running tests
go test -v ./pkg/... || true
