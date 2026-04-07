// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mcpServerProvider struct {
	server *mcp.Server
}

// Summary: Server executes the operation.
//
// Parameters:
//   - None
//
// Returns:
//   - *mcp.Server {
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (p *mcpServerProvider) Server() *mcp.Server {
	return p.server
}

// NewMCPServerProvider creates a new MCPServerProvider.
//
// Summary: Initializes a provider for the MCP server.
//
// Parameters:
// Summary: NewMCPServerProvider executes the operation.
//
// Parameters:
//   - server *mcp.Server: Input parameter.
//
// Returns:
//   - MCPServerProvider {
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func NewMCPServerProvider(server *mcp.Server) MCPServerProvider {
	return &mcpServerProvider{server: server}
}
