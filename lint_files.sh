#!/bin/bash
cd server
GOLANGCI_LINT_BIN=$(go env GOPATH)/bin/golangci-lint
for file in $(find . -name "*.go"); do
  $GOLANGCI_LINT_BIN run --concurrency 1 $file 2>/dev/null || true
done
