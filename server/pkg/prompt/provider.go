// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mcpServerProvider struct {
	server *mcp.Server
}

// Server provides server functionality.
//
// Summary: Server.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (p *mcpServerProvider) Server() *mcp.Server {
	return p.server
}

// NewMCPServerProvider provides newmcpserverprovider functionality.
//
// Summary: NewMCPServerProvider.
//
// Parameters.
//   - server: The parameter.
//
// Returns.
//   - result: The result.
func NewMCPServerProvider(server *mcp.Server) MCPServerProvider {
	return &mcpServerProvider{server: server}
}
