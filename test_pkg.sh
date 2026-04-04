#!/bin/bash
cd server
go mod edit -replace github.com/mcpany/core/proto=../proto
go mod tidy
go test -coverprofile=coverage.out ./pkg/util/schemaconv/...
go tool cover -func=coverage.out
