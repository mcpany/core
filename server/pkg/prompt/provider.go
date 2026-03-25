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
// Summary: Retrieves the MCP server.
//
// Returns: - None.
//   - *mcp.Server: The MCP server instance.
func (p *mcpServerProvider) Server() *mcp.Server {
	return p.server
}

// NewMCPServerProvider creates a new MCPServerProvider.
//
// Summary: Initializes a provider for the MCP server.
//
// Parameters: - None.
//   - server: *mcp.Server. The server instance to wrap.
//
// Returns: - None.
//   - MCPServerProvider: The initialized provider.
func NewMCPServerProvider(server *mcp.Server) MCPServerProvider {
	return &mcpServerProvider{server: server}
}
