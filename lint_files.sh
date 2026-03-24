#!/bin/bash
for file in $(find server/pkg -name "*.go"); do
  $(go env GOPATH)/bin/golangci-lint run --fast $file
done
