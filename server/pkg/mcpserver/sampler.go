// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"fmt"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Summary: MCPSession wraps an MCP session to provide client interaction capabilities like sampling and roots. Provides a wrapper around the MCP server session to facilitate client interactions.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type MCPSession struct {
	session *mcp.ServerSession
}

// Summary: NewMCPSession creates a new MCPSession. Initializes a new MCPSession instance.
//
// Parameters:
//   - session (*mcp.ServerSession): The session parameter.
//
// Returns:
//   - *MCPSession: The resulting *MCPSession.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewMCPSession(session *mcp.ServerSession) *MCPSession {
	return &MCPSession{session: session}
}

// Summary: NewMCPSampler is a deprecated alias for NewMCPSession. Creates a new MCPSession (deprecated alias).
//
// Parameters:
//   - session (*mcp.ServerSession): The session parameter.
//
// Returns:
//   - *MCPSession: The resulting *MCPSession.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewMCPSampler(session *mcp.ServerSession) *MCPSession {
	return NewMCPSession(session)
}

// Summary: CreateMessage requests a message creation from the client (sampling). Requests the client to create a message, effectively sampling the LLM.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//   - params (*mcp.CreateMessageParams): The params parameter.
//
// Returns:
//   - *mcp.CreateMessageResult: The resulting *mcp.CreateMessageResult.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *MCPSession) CreateMessage(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error) {
	if s.session == nil {
		return nil, fmt.Errorf("no active session available for sampling")
	}
	return s.session.CreateMessage(ctx, params)
}

// Summary: ListRoots requests the list of roots from the client. Requests the list of root directories from the client.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//
// Returns:
//   - *mcp.ListRootsResult: The resulting *mcp.ListRootsResult.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *MCPSession) ListRoots(ctx context.Context) (*mcp.ListRootsResult, error) {
	if s.session == nil {
		return nil, fmt.Errorf("no active session available for roots inspection")
	}
	// The SDK exposes ListRoots on ServerSession
	return s.session.ListRoots(ctx, nil)
}

// Verify that MCPSession implements tool.Session.
var _ tool.Session = (*MCPSession)(nil)
