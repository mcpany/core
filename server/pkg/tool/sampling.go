// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Session defines the interface for tools to interact with the client session.
//
// Summary: Interface for client session capabilities like Sampling and Roots inspection.
type Session interface {
	// CreateMessage requests a message creation (sampling) from the client.
	//
	// Summary: Requests the client to create a message (sampling).
	//
	// Parameters:
	//   - ctx: context.Context. The request context.
	//   - params: *mcp.CreateMessageParams. The parameters for the sampling request.
	//
	// Returns:
	//   - *mcp.CreateMessageResult: The result of the sampling request.
	//   - error: An error if the request fails.
	CreateMessage(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error)

	// ListRoots requests the list of roots from the client.
	//
	// Summary: Requests the client to list available filesystem roots.
	//
	// Parameters:
	//   - ctx: context.Context. The request context.
	//
	// Returns:
	//   - *mcp.ListRootsResult: The list of roots provided by the client.
	//   - error: An error if the request fails.
	ListRoots(ctx context.Context) (*mcp.ListRootsResult, error)
}

// Sampler is an alias for Session for backward compatibility.
type Sampler = Session

type sessionContextKey struct{}

// NewContextWithSession creates a new context with the given Session.
//
// Summary: Embeds a Session into the context.
//
// Parameters:
//   - ctx: context.Context. The parent context.
//   - s: Session. The session to embed.
//
// Returns:
//   - context.Context: The new context containing the session.
func NewContextWithSession(ctx context.Context, s Session) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, s)
}

// GetSession retrieves the Session from the context.
//
// Summary: Retrieves the Session from the context if present.
//
// Parameters:
//   - ctx: context.Context. The context to search.
//
// Returns:
//   - Session: The session instance.
//   - bool: True if the session was found, false otherwise.
func GetSession(ctx context.Context) (Session, bool) {
	s, ok := ctx.Value(sessionContextKey{}).(Session)
	return s, ok
}

// NewContextWithSampler creates a new context with the given Sampler.
//
// Deprecated: Use NewContextWithSession instead.
func NewContextWithSampler(ctx context.Context, s Sampler) context.Context {
	return NewContextWithSession(ctx, s)
}

// GetSampler retrieves the Sampler from the context.
//
// Deprecated: Use GetSession instead.
func GetSampler(ctx context.Context) (Sampler, bool) {
	return GetSession(ctx)
}
