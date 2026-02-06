// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mcpServerProvider struct {
	server *mcp.Server
}

// Server returns the underlying MCP server instance associated with this provider.
//
// Returns:
//   - *mcp.Server: The MCP server instance.
func (p *mcpServerProvider) Server() *mcp.Server {
	return p.server
}

// NewMCPServerProvider creates a new provider that wraps an MCP server.
//
// Parameters:
//   - server: *mcp.Server. The MCP server instance to wrap.
//
// Returns:
//   - MCPServerProvider: A new provider instance.
func NewMCPServerProvider(server *mcp.Server) MCPServerProvider {
	return &mcpServerProvider{server: server}
}
