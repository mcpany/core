#!/bin/bash
cd server
sed -i 's/go 1.26.1/go 1.24.0/' go.mod && sed -i 's/toolchain go1.26.1/toolchain go1.24.0/' go.mod
go mod tidy
~/go/bin/golangci-lint run ./pkg/... > lint_errors.txt || true
cat lint_errors.txt | grep check-go-doc | head -n 30
