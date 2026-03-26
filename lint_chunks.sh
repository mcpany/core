#!/bin/bash
cd server
for dir in cmd pkg tests examples; do
  echo "Linting $dir..."
  $(go env GOPATH)/bin/golangci-lint run --concurrency 1 --timeout 10m ./$dir/...
done
