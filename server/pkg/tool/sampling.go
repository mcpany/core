// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Session defines the interface for tools to interact with the client session.
// It includes capabilities like Sampling (CreateMessage) and Roots inspection.
type Session interface {
	// CreateMessage requests a message creation (sampling) from the client.
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

// NewContextWithSession creates a new context with the given Session.
//
// Parameters:
//   - ctx: context.Context. The context to extend.
//   - s: Session. The session to embed.
//
// Returns:
//   - context.Context: A new context containing the session.
func NewContextWithSession(ctx context.Context, s Session) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, s)
}

// GetSession retrieves the Session from the context.
//
// Parameters:
//   - ctx: context.Context. The context to search.
//
// Returns:
//   - Session: The session if found.
//   - bool: True if the session was found.
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
