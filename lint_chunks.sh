#!/bin/bash
export GOMEMLIMIT=2000MiB
export PATH=$PATH:$(go env GOPATH)/bin
cd server

echo "Linting pkg/middleware..."
golangci-lint run -c .golangci.yml -j 2 --timeout 20m --fix ./pkg/middleware/...

echo "Linting cmd..."
golangci-lint run -c .golangci.yml -j 2 --timeout 20m --fix ./cmd/...

echo "Linting tests..."
golangci-lint run -c .golangci.yml -j 2 --timeout 20m --fix ./tests/...

echo "Linting pkg (minus middleware)..."
# Skip middleware since we already linted it, or just lint all of pkg.
# For CI, it's probably better to just lint specific packages if it's OOMing.
# But let's see if pkg/... passes alone.
golangci-lint run -c .golangci.yml -j 2 --timeout 20m --fix ./pkg/...

echo "Done"
