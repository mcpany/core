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

// NewMCPSession initializes a new MCPSession wrapper.
//
// Parameters:
//   - session: The underlying MCP server session.
//
// Returns:
//   - *MCPSession: The initialized session wrapper.
//
// Side Effects:
//   - None.
func NewMCPSession(session *mcp.ServerSession) *MCPSession {
	return &MCPSession{session: session}
}

// NewMCPSampler creates a new MCPSession (deprecated alias).
//
// Parameters:
//   - session: The underlying MCP server session.
//
// Returns:
//   - *MCPSession: The initialized session wrapper.
//
// Side Effects:
//   - None.
func NewMCPSampler(session *mcp.ServerSession) *MCPSession {
	return NewMCPSession(session)
}

// CreateMessage requests the client to generate a message (sampling).
//
// Parameters:
//   - ctx: Request context.
//   - params: Parameters for the sampling request (messages, model preferences).
//
// Returns:
//   - *mcp.CreateMessageResult: The generated message from the client.
//   - error: If the session is nil or the client returns an error.
//
// Errors:
//   - Returns error if no active session is available.
//   - Returns error if the client fails to generate a message.
//
// Side Effects:
//   - Sends a JSON-RPC request to the connected client.
func (s *MCPSession) CreateMessage(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error) {
	if s.session == nil {
		return nil, fmt.Errorf("no active session available for sampling")
	}
	return s.session.CreateMessage(ctx, params)
}

// ListRoots requests the list of roots from the client.
//
// Parameters:
//   - ctx: Request context.
//
// Returns:
//   - *mcp.ListRootsResult: The list of roots provided by the client.
//   - error: If the session is nil or the client returns an error.
//
// Errors:
//   - Returns error if no active session is available.
//   - Returns error if the client fails to list roots.
//
// Side Effects:
//   - Sends a JSON-RPC request to the connected client.
func (s *MCPSession) ListRoots(ctx context.Context) (*mcp.ListRootsResult, error) {
	if s.session == nil {
		return nil, fmt.Errorf("no active session available for roots inspection")
	}
	// The SDK exposes ListRoots on ServerSession
	return s.session.ListRoots(ctx, nil)
}

// Verify that MCPSession implements tool.Session.
var _ tool.Session = (*MCPSession)(nil)
