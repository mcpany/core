package mcpserver

import (
	"context"
	"fmt"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPSession wraps an MCP session to provide client interaction capabilities like sampling and roots.
//
// Summary: Provides a wrapper around the MCP server session to facilitate client interactions.
type MCPSession struct {
	session *mcp.ServerSession
}

// NewMCPSession creates a new MCPSession.
//
// Summary: Initializes a new MCPSession instance.
//
// Parameters:
//   - session: *mcp.ServerSession. The underlying MCP server session.
//
// Returns:
//   - *MCPSession: A new instance of MCPSession.
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
//
// Throws/Errors:
//   - Returns an error if the session is nil.
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
func (s *MCPSession) ListRoots(ctx context.Context) (*mcp.ListRootsResult, error) {
	if s.session == nil {
		return nil, fmt.Errorf("no active session available for roots inspection")
	}
	// The SDK exposes ListRoots on ServerSession
	return s.session.ListRoots(ctx, nil)
}

// Verify that MCPSession implements tool.Session.
var _ tool.Session = (*MCPSession)(nil)
