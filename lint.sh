#!/bin/bash
set -e
export PATH=$PATH:$(go env GOPATH)/bin
cd server
sed -i 's/go 1.26.1/go 1.24/g' go.mod
sed -i 's/toolchain go1.26.1/toolchain go1.24.0/g' go.mod
cd ..
sed -i 's/go 1.26.1/go 1.24/g' go.work
golangci-lint run --config server/.golangci.yml ./server/...
git restore server/go.mod go.work
