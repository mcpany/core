// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mcpServerProvider struct {
	server *mcp.Server
}

// Server serves as a public interface for interacting with Server.
//
// Summary: Server the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (p *mcpServerProvider) Server() *mcp.Server {
	return p.server
}

// NewMCPServerProvider serves as a public interface for interacting with NewMCPServerProvider.
//
// Summary: Constructs and returns an initialized mcp server provider ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewMCPServerProvider(server *mcp.Server) MCPServerProvider {
	return &mcpServerProvider{server: server}
}
