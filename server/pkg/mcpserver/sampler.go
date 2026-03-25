// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver
// NewMCPSampler is a deprecated alias for NewMCPSession.
// CreateMessage requests a message creation from the client (sampling).
//
// Summary: Requests the client to create a message, effectively sampling the LLM.
//
// Parameters:
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
//   - ctx: context.Context. The context for the request.
//   - params: *mcp.CreateMessageParams. The parameters for the message creation request.
// ListRoots requests the list of roots from the client.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Summary: Requests the list of root directories from the client.
//
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Parameters:
//   - ctx: context.Context. The context for the request.
//
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Returns:
//   - *mcp.ListRootsResult: The list of roots returned by the client.
//   - error: An error if no active session is available or if the request fails.
//
// Throws/Errors:
//   - Returns an error if the session is nil.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (s *MCPSession) ListRoots(ctx context.Context) (*mcp.ListRootsResult, error) {
	if s.session == nil {
		return nil, fmt.Errorf("no active session available for roots inspection")
	}
	// The SDK exposes ListRoots on ServerSession
	return s.session.ListRoots(ctx, nil)
}

// Verify that MCPSession implements tool.Session.
var _ tool.Session = (*MCPSession)(nil)
