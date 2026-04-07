#!/bin/bash
if ! grep -q '"github.com/mcpany/core/server/pkg/middleware"' server/pkg/app/api.go; then
    sed -i '/"github.com\/modelcontextprotocol\/go-sdk\/mcp"/a \t"github.com/mcpany/core/server/pkg/middleware"' server/pkg/app/api.go
fi
