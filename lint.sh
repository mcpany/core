#!/bin/bash
set -e
export GOPATH=$(go env GOPATH)
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $GOPATH/bin v1.64.6
$GOPATH/bin/golangci-lint run ./server/pkg/app/...
