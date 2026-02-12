// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Session defines the interface for tools to interact with the client session.
// It includes capabilities like Sampling (CreateMessage) and Roots inspection.
//
// Summary: Interface for client session interactions.
type Session interface {
	// CreateMessage requests a message creation (sampling) from the client.
	//
	// Summary: Requests message creation from the client.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - params: *mcp.CreateMessageParams. The parameters for the message creation.
	//
	// Returns:
	//   - *mcp.CreateMessageResult: The result of the message creation.
	//   - error: An error if the operation fails.
	CreateMessage(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error)

	// ListRoots requests the list of roots from the client.
	//
	// Summary: Requests the list of roots from the client.
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
//
// Summary: Alias for Session (Deprecated).
type Sampler = Session

type sessionContextKey struct{}

// NewContextWithSession creates a new context with the given Session.
//
// Summary: Embeds a Session into the context.
//
// Parameters:
//   - ctx: context.Context. The context to extend.
//   - s: Session. The session to embed.
//
// Returns:
//   - context.Context: The new context.
func NewContextWithSession(ctx context.Context, s Session) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, s)
}

// GetSession retrieves the Session from the context.
//
// Summary: Retrieves the Session from the context.
//
// Parameters:
//   - ctx: context.Context. The context to search.
//
// Returns:
//   - Session: The session if found.
//   - bool: True if found, false otherwise.
func GetSession(ctx context.Context) (Session, bool) {
	s, ok := ctx.Value(sessionContextKey{}).(Session)
	return s, ok
}

// NewContextWithSampler creates a new context with the given Sampler.
//
// Summary: Embeds a Sampler into the context (Deprecated).
//
// Parameters:
//   - ctx: context.Context. The context to extend.
//   - s: Sampler. The sampler to embed.
//
// Returns:
//   - context.Context: The new context.
//
// Deprecated: Use NewContextWithSession instead.
func NewContextWithSampler(ctx context.Context, s Sampler) context.Context {
	return NewContextWithSession(ctx, s)
}

// GetSampler retrieves the Sampler from the context.
//
// Summary: Retrieves the Sampler from the context (Deprecated).
//
// Parameters:
//   - ctx: context.Context. The context to search.
//
// Returns:
//   - Sampler: The sampler if found.
//   - bool: True if found, false otherwise.
//
// Deprecated: Use GetSession instead.
func GetSampler(ctx context.Context) (Sampler, bool) {
	return GetSession(ctx)
}
