#!/bin/bash
export GOTOOLCHAIN=local
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.6
~/go/bin/golangci-lint run > lint.out || true
cat lint.out | grep "revive"
