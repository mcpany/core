// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mcpServerProvider struct {
	server *mcp.Server
}

// Summary: Server returns the underlying MCP server instance. Retrieves the MCP server.
//
// Parameters:
//   - None.
//
// Returns:
//   - *mcp.Server: The resulting *mcp.Server.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (p *mcpServerProvider) Server() *mcp.Server {
	return p.server
}

// Summary: NewMCPServerProvider creates a new MCPServerProvider. Initializes a provider for the MCP server.
//
// Parameters:
//   - server (*mcp.Server): The server parameter.
//
// Returns:
//   - MCPServerProvider: The resulting MCPServerProvider.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewMCPServerProvider(server *mcp.Server) MCPServerProvider {
	return &mcpServerProvider{server: server}
}
