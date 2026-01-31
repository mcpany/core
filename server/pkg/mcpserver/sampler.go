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
// Parameters:
//   - session: *mcp.ServerSession. The underlying MCP session.
//
// Returns:
//   - *MCPSession: The wrapped session.
func NewMCPSession(session *mcp.ServerSession) *MCPSession {
	return &MCPSession{session: session}
}

// NewMCPSampler is a deprecated alias for NewMCPSession.
//
// Parameters:
//   - session: *mcp.ServerSession. The underlying MCP session.
//
// Returns:
//   - *MCPSession: The wrapped session.
func NewMCPSampler(session *mcp.ServerSession) *MCPSession {
	return NewMCPSession(session)
}

// CreateMessage requests a message creation from the client (sampling).
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - params: *mcp.CreateMessageParams. The parameters for generating the message.
//
// Returns:
//   - *mcp.CreateMessageResult: The generated message.
//   - error: An error if the session is invalid or the request fails.
func (s *MCPSession) CreateMessage(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error) {
	if s.session == nil {
		return nil, fmt.Errorf("no active session available for sampling")
	}
	return s.session.CreateMessage(ctx, params)
}

// ListRoots requests the list of roots from the client.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//
// Returns:
//   - *mcp.ListRootsResult: The list of roots provided by the client.
//   - error: An error if the session is invalid or the request fails.
func (s *MCPSession) ListRoots(ctx context.Context) (*mcp.ListRootsResult, error) {
	if s.session == nil {
		return nil, fmt.Errorf("no active session available for roots inspection")
	}
	// The SDK exposes ListRoots on ServerSession
	return s.session.ListRoots(ctx, nil)
}

// Verify that MCPSession implements tool.Session.
var _ tool.Session = (*MCPSession)(nil)
