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
// Summary: Returns the underlying MCP server instance.
//
// Parameters:
//   - None.
//
// Returns:
//   - *mcp.Server: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (p *mcpServerProvider) Server() *mcp.Server {
	return p.server
}

// NewMCPServerProvider creates a new MCPServerProvider.
//
// Summary: Creates a new MCPServerProvider.
//
// Parameters:
//   - server (*mcp.Server): Parameter.
//
// Returns:
//   - MCPServerProvider: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func NewMCPServerProvider(server *mcp.Server) MCPServerProvider {
	return &mcpServerProvider{server: server}
}
