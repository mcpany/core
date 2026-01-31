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
// Returns:
//   - *mcp.Server: The MCP server instance.
func (p *mcpServerProvider) Server() *mcp.Server {
	return p.server
}

// NewMCPServerProvider creates a new MCPServerProvider.
//
// Parameters:
//   - server: *mcp.Server. The MCP server instance.
//
// Returns:
//   - MCPServerProvider: The provider instance.
func NewMCPServerProvider(server *mcp.Server) MCPServerProvider {
	return &mcpServerProvider{server: server}
}
