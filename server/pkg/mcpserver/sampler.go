// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"fmt"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPSession wraps an MCP session to provide client interaction capabilities like sampling and roots. Summary: Provides a wrapper around the MCP server session to facilitate client interactions.
//
// Summary: MCPSession wraps an MCP session to provide client interaction capabilities like sampling and roots. Summary: Provides a wrapper around the MCP server session to facilitate client interactions.
//
// Fields:
//   - Contains the configuration and state properties required for MCPSession functionality.
type MCPSession struct {
	session *mcp.ServerSession
}

// NewMCPSession creates a new MCPSession. Summary: Initializes a new MCPSession instance. Parameters: - session: *mcp.ServerSession. The underlying MCP server session. Returns: - *MCPSession: A new instance of MCPSession.
//
// Summary: NewMCPSession creates a new MCPSession. Summary: Initializes a new MCPSession instance. Parameters: - session: *mcp.ServerSession. The underlying MCP server session. Returns: - *MCPSession: A new instance of MCPSession.
//
// Parameters:
//   - session (*mcp.ServerSession): The session parameter used in the operation.
//
// Returns:
//   - (*MCPSession): The resulting MCPSession object containing the requested data.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
//   - *MCPSession: A new instance of MCPSession.
//
// Side Effects:
//   - This function is deprecated and should be replaced by NewMCPSession.
func NewMCPSampler(session *mcp.ServerSession) *MCPSession {
	return NewMCPSession(session)
}

// CreateMessage requests a message creation from the client (sampling). Summary: Requests the client to create a message, effectively sampling the LLM. Parameters: - ctx: context.Context. The context for the request. - params: *mcp.CreateMessageParams. The parameters for the message creation request. Returns: - *mcp.CreateMessageResult: The result of the message creation from the client. - error: An error if no active session is available or if the request fails. Throws/Errors: - Returns an error if the session is nil.
//
// Summary: CreateMessage requests a message creation from the client (sampling). Summary: Requests the client to create a message, effectively sampling the LLM. Parameters: - ctx: context.Context. The context for the request. - params: *mcp.CreateMessageParams. The parameters for the message creation request. Returns: - *mcp.CreateMessageResult: The result of the message creation from the client. - error: An error if no active session is available or if the request fails. Throws/Errors: - Returns an error if the session is nil.
//
// Parameters:
//   - ctx (context.Context): The context for managing request lifecycle and cancellation.
//   - params (*mcp.CreateMessageParams): The params parameter used in the operation.
//
// Returns:
//   - (*mcp.CreateMessageResult): The resulting mcp.CreateMessageResult object containing the requested data.
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
//
// Side Effects:
//   - Modifies global state, writes to the database, or establishes network connections.
func (s *MCPSession) CreateMessage(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error) {
	if s.session == nil {
		return nil, fmt.Errorf("no active session available for sampling")
	}
	return s.session.CreateMessage(ctx, params)
}

// ListRoots requests the list of roots from the client. Summary: Requests the list of root directories from the client. Parameters: - ctx: context.Context. The context for the request. Returns: - *mcp.ListRootsResult: The list of roots returned by the client. - error: An error if no active session is available or if the request fails. Throws/Errors: - Returns an error if the session is nil.
//
// Summary: ListRoots requests the list of roots from the client. Summary: Requests the list of root directories from the client. Parameters: - ctx: context.Context. The context for the request. Returns: - *mcp.ListRootsResult: The list of roots returned by the client. - error: An error if no active session is available or if the request fails. Throws/Errors: - Returns an error if the session is nil.
//
// Parameters:
//   - ctx (context.Context): The context for managing request lifecycle and cancellation.
//
// Returns:
//   - (*mcp.ListRootsResult): The resulting mcp.ListRootsResult object containing the requested data.
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
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
