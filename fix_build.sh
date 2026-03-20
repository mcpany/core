#!/bin/bash
cd server
go mod edit -dropreplace github.com/mcpany/core/proto
go get github.com/mcpany/core/proto@latest
go mod tidy
make build
