// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"fmt"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPSession wraps an MCP session to provide client interaction capabilities like sampling and roots.
type MCPSession struct {
	session *mcp.ServerSession
}

// NewMCPSession creates a new MCPSession.
//
// session is the session.
//
// Returns the result.
//
// Parameters:
//   - session: *mcp.ServerSession. The mcp.ServerSession instance.
//
// Returns:
//   - *MCPSession: The result.
func NewMCPSession(session *mcp.ServerSession) *MCPSession {
	return &MCPSession{session: session}
}

// NewMCPSampler is a deprecated alias for NewMCPSession.
//
// session is the session.
//
// Returns the result.
//
// Parameters:
//   - session: *mcp.ServerSession. The mcp.ServerSession instance.
//
// Returns:
//   - *MCPSession: The result.
func NewMCPSampler(session *mcp.ServerSession) *MCPSession {
	return NewMCPSession(session)
}

// CreateMessage requests a message creation from the client (sampling).
//
// ctx is the context for the request.
// params is the params.
//
// Returns the result.
// Returns an error if the operation fails.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - params: *mcp.CreateMessageParams. The mcp.CreateMessageParams instance.
//
// Returns:
//   - *mcp.CreateMessageResult: The result.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   - Returns an error if the operation fails.
func (s *MCPSession) CreateMessage(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error) {
	if s.session == nil {
		return nil, fmt.Errorf("no active session available for sampling")
	}
	return s.session.CreateMessage(ctx, params)
}

// ListRoots requests the list of roots from the client.
//
// ctx is the context for the request.
//
// Returns the result.
// Returns an error if the operation fails.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//
// Returns:
//   - *mcp.ListRootsResult: The result.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   - Returns an error if the operation fails.
func (s *MCPSession) ListRoots(ctx context.Context) (*mcp.ListRootsResult, error) {
	if s.session == nil {
		return nil, fmt.Errorf("no active session available for roots inspection")
	}
	// The SDK exposes ListRoots on ServerSession
	return s.session.ListRoots(ctx, nil)
}

// Verify that MCPSession implements tool.Session.
var _ tool.Session = (*MCPSession)(nil)
