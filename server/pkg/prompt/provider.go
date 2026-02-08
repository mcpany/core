// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mcpServerProvider struct {
	server *mcp.Server
}

// Server returns the underlying MCP server instance.
//
// Summary: returns the underlying MCP server instance.
//
// Parameters:
//   None.
//
// Returns:
//   - *mcp.Server: The *mcp.Server.
func (p *mcpServerProvider) Server() *mcp.Server {
	return p.server
}

// NewMCPServerProvider creates a new MCPServerProvider.
//
// Summary: creates a new MCPServerProvider.
//
// Parameters:
//   - server: *mcp.Server. The server.
//
// Returns:
//   - MCPServerProvider: The MCPServerProvider.
func NewMCPServerProvider(server *mcp.Server) MCPServerProvider {
	return &mcpServerProvider{server: server}
}
