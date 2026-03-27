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
//
// Summary. Provides a wrapper around the MCP server session to facilitate client interactions.
type MCPSession struct {
	session *mcp.ServerSession
}

// NewMCPSession provides newmcpsession functionality.
//
// Summary: NewMCPSession.
//
// Parameters.
//   - session: The parameter.
//
// Returns.
//   - result: The result.
func NewMCPSession(session *mcp.ServerSession) *MCPSession {
	return &MCPSession{session: session}
}

// NewMCPSampler provides newmcpsampler functionality.
//
// Summary: NewMCPSampler.
//
// Parameters.
//   - session: The parameter.
//
// Returns.
//   - result: The result.
func NewMCPSampler(session *mcp.ServerSession) *MCPSession {
	return NewMCPSession(session)
}

// CreateMessage provides createmessage functionality.
//
// Summary: CreateMessage.
//
// Parameters.
//   - ctx: The parameter.
//   - params: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (s *MCPSession) CreateMessage(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error) {
	if s.session == nil {
		return nil, fmt.Errorf("no active session available for sampling")
	}
	return s.session.CreateMessage(ctx, params)
}

// ListRoots provides listroots functionality.
//
// Summary: ListRoots.
//
// Parameters.
//   - ctx: The parameter.
//   - error: The parameter.
//
// Returns.
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
