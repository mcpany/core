// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"fmt"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPSession represents the public MCPSession entity.
//
// Summary: Defines the structured data model representing a session.
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

// NewMCPSession serves as a public interface for interacting with NewMCPSession.
//
// Summary: Constructs and returns an initialized mcp session ready for consumption.
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
func NewMCPSession(session *mcp.ServerSession) *MCPSession {
	return &MCPSession{session: session}
}

// NewMCPSampler serves as a public interface for interacting with NewMCPSampler.
//
// Summary: Constructs and returns an initialized mcp sampler ready for consumption.
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
func NewMCPSampler(session *mcp.ServerSession) *MCPSession {
	return NewMCPSession(session)
}

// CreateMessage serves as a public interface for interacting with CreateMessage.
//
// Summary: Create the message appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *MCPSession) CreateMessage(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error) {
	if s.session == nil {
		return nil, fmt.Errorf("no active session available for sampling")
	}
	return s.session.CreateMessage(ctx, params)
}

// ListRoots serves as a public interface for interacting with ListRoots.
//
// Summary: List the roots appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *MCPSession) ListRoots(ctx context.Context) (*mcp.ListRootsResult, error) {
	if s.session == nil {
		return nil, fmt.Errorf("no active session available for roots inspection")
	}
	// The SDK exposes ListRoots on ServerSession
	return s.session.ListRoots(ctx, nil)
}

// Verify that MCPSession implements tool.Session.
var _ tool.Session = (*MCPSession)(nil)
