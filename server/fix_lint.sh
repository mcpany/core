#!/bin/bash
export PATH=$PATH:~/.local/bin
export GO111MODULE=on
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.6
$(go env GOPATH)/bin/golangci-lint run ./pkg/app/...
