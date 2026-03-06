// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Session - Auto-generated documentation.
//
// Summary: Session defines the interface for tools to interact with the client session.
//
// Methods:
//   - Various methods for Session.
type Session interface {
	// CreateMessage requests a message creation (sampling) from the client.
	//
	// Summary: Requests message creation.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - params: *mcp.CreateMessageParams. The parameters for message creation.
	//
	// Returns:
	//   - *mcp.CreateMessageResult: The result of the message creation.
	//   - error: An error if the operation fails.
	CreateMessage(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error)

	// ListRoots requests the list of roots from the client.
	//
	// Summary: Requests roots list.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//
	// Returns:
	//   - *mcp.ListRootsResult: The list of roots.
	//   - error: An error if the operation fails.
	ListRoots(ctx context.Context) (*mcp.ListRootsResult, error)
}

// Sampler is an alias for Session for backward compatibility.
type Sampler = Session

type sessionContextKey struct{}

// NewContextWithSession creates a new context with the given Session. Summary: Injects Session into context. Parameters: - ctx: context.Context. The parent context. - s: Session. The session to inject. Returns: - context.Context: The new context.
//
// Summary: NewContextWithSession creates a new context with the given Session. Summary: Injects Session into context. Parameters: - ctx: context.Context. The parent context. - s: Session. The session to inject. Returns: - context.Context: The new context.
//
// Parameters:
//   - ctx (context.Context): The context for managing request lifecycle and cancellation.
//   - s (Session): The s parameter used in the operation.
//
// Returns:
//   - (context.Context): The resulting context.Context object containing the requested data.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewContextWithSession(ctx context.Context, s Session) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, s)
}

// GetSession retrieves the Session from the context. Summary: Retrieves Session from context. Parameters: - ctx: context.Context. The context. Returns: - Session: The session if found. - bool: True if the session exists.
//
// Summary: GetSession retrieves the Session from the context. Summary: Retrieves Session from context. Parameters: - ctx: context.Context. The context. Returns: - Session: The session if found. - bool: True if the session exists.
//
// Parameters:
//   - ctx (context.Context): The context for managing request lifecycle and cancellation.
//
// Returns:
//   - (Session): The resulting Session object containing the requested data.
//   - (bool): A boolean indicating the success or status of the operation.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func GetSession(ctx context.Context) (Session, bool) {
	s, ok := ctx.Value(sessionContextKey{}).(Session)
	return s, ok
}

// NewContextWithSampler creates a new context with the given Sampler. Summary: Injects Sampler into context. Deprecated: Use NewContextWithSession instead. Parameters: - ctx: context.Context. The parent context. - s: Sampler. The sampler to inject. Returns: - context.Context: The new context.
//
// Summary: NewContextWithSampler creates a new context with the given Sampler. Summary: Injects Sampler into context. Deprecated: Use NewContextWithSession instead. Parameters: - ctx: context.Context. The parent context. - s: Sampler. The sampler to inject. Returns: - context.Context: The new context.
//
// Parameters:
//   - ctx (context.Context): The context for managing request lifecycle and cancellation.
//   - s (Sampler): The s parameter used in the operation.
//
// Returns:
//   - (context.Context): The resulting context.Context object containing the requested data.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewContextWithSampler(ctx context.Context, s Sampler) context.Context {
	return NewContextWithSession(ctx, s)
}

// GetSampler retrieves the Sampler from the context. Summary: Retrieves Sampler from context. Deprecated: Use GetSession instead. Parameters: - ctx: context.Context. The context. Returns: - Sampler: The sampler if found. - bool: True if the sampler exists.
//
// Summary: GetSampler retrieves the Sampler from the context. Summary: Retrieves Sampler from context. Deprecated: Use GetSession instead. Parameters: - ctx: context.Context. The context. Returns: - Sampler: The sampler if found. - bool: True if the sampler exists.
//
// Parameters:
//   - ctx (context.Context): The context for managing request lifecycle and cancellation.
//
// Returns:
//   - (Sampler): The resulting Sampler object containing the requested data.
//   - (bool): A boolean indicating the success or status of the operation.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func GetSampler(ctx context.Context) (Sampler, bool) {
	return GetSession(ctx)
}
