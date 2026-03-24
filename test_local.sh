#!/bin/bash
export PATH=$PATH:$(go env GOPATH)/bin
sed -i 's/go 1.26.1/go 1.24/g' go.mod
sed -i 's/toolchain go1.26.1/toolchain go1.24.0/g' go.mod
cd server
sed -i 's/go 1.26.1/go 1.24/g' go.mod
golangci-lint run --tests=false ./pkg/storage/postgres/...
git checkout go.mod
cd ..
git checkout go.mod
