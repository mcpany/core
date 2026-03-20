#!/bin/bash
export PATH=/tmp/go1.26/bin:$PATH
export GOTOOLCHAIN=auto
cd server
sed -i 's/go 1.24.6/go 1.24/g' go.work
sed -i 's/go 1.26.1/go 1.24/g' go.mod
sed -i 's/go: "1.26.1"/go: "1.24"/g' .golangci.yml
$(go env GOPATH)/bin/golangci-lint run --disable-all -E revive -E gofumpt -E errcheck -E wastedassign ./cmd/server/...
git restore go.mod .golangci.yml ../go.work
