// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"fmt"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Summary: MCPSession represents a data structure.
//
// Parameters:
//   - None
//
// Returns:
//   - None
//
// Errors:
//   - None
//
// Side Effects:
//   - None
type MCPSession struct {
	session *mcp.ServerSession
}

// NewMCPSession creates a new MCPSession.
//
// Summary: NewMCPSession executes the operation.
//
// Parameters:
//   - session *mcp.ServerSession: Input parameter.
//
// Returns:
//   - *MCPSession {
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func NewMCPSession(session *mcp.ServerSession) *MCPSession {
	return &MCPSession{session: session}
}

// NewMCPSampler is a deprecated alias for NewMCPSession.
//
// Summary: Creates a new MCPSession (deprecated alias).
//
// Parameters:
//   - session: *mcp.ServerSession. The underlying MCP server session.
//
// Returns:
// Summary: NewMCPSampler executes the operation.
//
// Parameters:
//   - session *mcp.ServerSession: Input parameter.
//
// Returns:
//   - *MCPSession {
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func NewMCPSampler(session *mcp.ServerSession) *MCPSession {
	return NewMCPSession(session)
}

// CreateMessage requests a message creation from the client (sampling).
//
// Summary: Requests the client to create a message, effectively sampling the LLM.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - params: *mcp.CreateMessageParams. The parameters for the message creation request.
//
// Returns:
//   - *mcp.CreateMessageResult: The result of the message creation from the client.
//   - error: An error if no active session is available or if the request fails.
// Summary: CreateMessage executes the operation.
//
// Parameters:
//   - ctx context.Context: Input parameter.
//   - params *mcp.CreateMessageParams: Input parameter.
//
// Returns:
//   - (*mcp.CreateMessageResult, error): Result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None
func (s *MCPSession) CreateMessage(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error) {
	if s.session == nil {
		return nil, fmt.Errorf("no active session available for sampling")
	}
	return s.session.CreateMessage(ctx, params)
}

// ListRoots requests the list of roots from the client.
//
// Summary: Requests the list of root directories from the client.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//
// Returns:
//   - *mcp.ListRootsResult: The list of roots returned by the client.
//   - error: An error if no active session is available or if the request fails.
//
// Throws/Errors:
//   - Returns an error if the session is nil.
// Summary: ListRoots executes the operation.
//
// Parameters:
//   - ctx context.Context: Input parameter.
//
// Returns:
//   - (*mcp.ListRootsResult, error): Result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None
func (s *MCPSession) ListRoots(ctx context.Context) (*mcp.ListRootsResult, error) {
	if s.session == nil {
		return nil, fmt.Errorf("no active session available for roots inspection")
	}
	// The SDK exposes ListRoots on ServerSession
	return s.session.ListRoots(ctx, nil)
}

// Verify that MCPSession implements tool.Session.
var _ tool.Session = (*MCPSession)(nil)
