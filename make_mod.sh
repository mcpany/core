#!/bin/bash
cd server
go mod edit -replace github.com/mcpany/core=../
go mod edit -replace github.com/mcpany/core/proto=../proto
go mod tidy
