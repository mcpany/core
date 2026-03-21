#!/bin/bash
export PATH=$PATH:$(go env GOPATH)/bin
cd server
golangci-lint run --timeout 20m --fix ./pkg/middleware/...
