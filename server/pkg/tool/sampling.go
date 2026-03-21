// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Summary: Session defines the interface for tools to interact with the client session. It includes capabilities like Sampling (CreateMessage) and Roots inspection. Represents a Session.
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

// Summary: Sampler is an alias for Session for backward compatibility. Represents a Sampler.
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
type Sampler = Session

type sessionContextKey struct{}

// Summary: NewContextWithSession creates a new context with the given Session. Injects Session into context.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//   - s (Session): The s parameter.
//
// Returns:
//   - context.Context: The resulting context.Context.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewContextWithSession(ctx context.Context, s Session) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, s)
}

// Summary: GetSession retrieves the Session from the context. Retrieves Session from context.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//
// Returns:
//   - Session: The resulting Session.
//   - bool: The resulting bool.
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

// Summary: NewContextWithSampler creates a new context with the given Sampler. Injects Sampler into context. Deprecated: Use NewContextWithSession instead.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//   - s (Sampler): The s parameter.
//
// Returns:
//   - context.Context: The resulting context.Context.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewContextWithSampler(ctx context.Context, s Sampler) context.Context {
	return NewContextWithSession(ctx, s)
}

// Summary: GetSampler retrieves the Sampler from the context. Retrieves Sampler from context. Deprecated: Use GetSession instead.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//
// Returns:
//   - Sampler: The resulting Sampler.
//   - bool: The resulting bool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func GetSampler(ctx context.Context) (Sampler, bool) {
	return GetSession(ctx)
}
