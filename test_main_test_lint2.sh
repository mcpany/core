#!/bin/bash
export PATH=/tmp/go1.26/bin:$PATH
export GOTOOLCHAIN=local
cd server
go mod edit -go=1.24.0 go.mod
go work edit -go=1.24.0 ../go.work
go work use .
go install github.com/mgechev/revive@latest
$(go env GOPATH)/bin/revive -config .golangci-revive.toml -formatter friendly ./cmd/server/... || $(go env GOPATH)/bin/revive -formatter friendly ./cmd/server/...
git checkout ../go.work go.mod
