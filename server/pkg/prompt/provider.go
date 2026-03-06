// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mcpServerProvider struct {
	server *mcp.Server
}

// Server - Auto-generated documentation.
//
// Summary: Server returns the underlying MCP server instance.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (p *mcpServerProvider) Server() *mcp.Server {
	return p.server
}

// NewMCPServerProvider creates a new MCPServerProvider. Summary: Initializes a provider for the MCP server. Parameters: - server: *mcp.Server. The server instance to wrap. Returns: - MCPServerProvider: The initialized provider.
//
// Summary: NewMCPServerProvider creates a new MCPServerProvider. Summary: Initializes a provider for the MCP server. Parameters: - server: *mcp.Server. The server instance to wrap. Returns: - MCPServerProvider: The initialized provider.
//
// Parameters:
//   - server (*mcp.Server): The server parameter used in the operation.
//
// Returns:
//   - (MCPServerProvider): The resulting MCPServerProvider object containing the requested data.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewMCPServerProvider(server *mcp.Server) MCPServerProvider {
	return &mcpServerProvider{server: server}
}
