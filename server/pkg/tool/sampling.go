// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Session defines the interface for tools to interact with the client session.
//
// Summary: defines the interface for tools to interact with the client session.
type Session interface {
	// CreateMessage requests a message creation (sampling) from the client.
	//
	// Summary: requests a message creation (sampling) from the client.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the operation.
	//   - params: *mcp.CreateMessageParams. The create message params.
	//
	// Returns:
	//   - *mcp.CreateMessageResult: The *mcp.CreateMessageResult.
	//   - error: An error if the operation fails.
	//
	// Throws/Errors:
	//   Returns an error if the operation fails.
	CreateMessage(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error)

	// ListRoots requests the list of roots from the client.
	//
	// Summary: requests the list of roots from the client.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the operation.
	//
	// Returns:
	//   - *mcp.ListRootsResult: The *mcp.ListRootsResult.
	//   - error: An error if the operation fails.
	//
	// Throws/Errors:
	//   Returns an error if the operation fails.
	ListRoots(ctx context.Context) (*mcp.ListRootsResult, error)
}

// Sampler is an alias for Session for backward compatibility.
//
// Summary: is an alias for Session for backward compatibility.
type Sampler = Session

type sessionContextKey struct{}

// NewContextWithSession creates a new context with the given Session.
//
// Summary: creates a new context with the given Session.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - s: Session. The s.
//
// Returns:
//   - context.Context: The context.Context.
func NewContextWithSession(ctx context.Context, s Session) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, s)
}

// GetSession retrieves the Session from the context.
//
// Summary: retrieves the Session from the context.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//
// Returns:
//   - Session: The Session.
//   - bool: The bool.
func GetSession(ctx context.Context) (Session, bool) {
	s, ok := ctx.Value(sessionContextKey{}).(Session)
	return s, ok
}

// NewContextWithSampler creates a new context with the given Sampler.
//
// Summary: creates a new context with the given Sampler.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - s: Sampler. The s.
//
// Returns:
//   - context.Context: The context.Context.
func NewContextWithSampler(ctx context.Context, s Sampler) context.Context {
	return NewContextWithSession(ctx, s)
}

// GetSampler retrieves the Sampler from the context.
//
// Summary: retrieves the Sampler from the context.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//
// Returns:
//   - Sampler: The Sampler.
//   - bool: The bool.
func GetSampler(ctx context.Context) (Sampler, bool) {
	return GetSession(ctx)
}
