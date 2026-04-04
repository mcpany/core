// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Session defines the interface for tools to interact with the client session.
//
// Summary: Defines the interface for tools to interact with the client session.
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

// Sampler is an alias for Session for backward compatibility.
//
// Summary: Is an alias for Session for backward compatibility.
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

// NewContextWithSession creates a new context with the given Session.
//
// Summary: Creates a new context with the given Session.
//
// Parameters:
//   - ctx (context.Context): Parameter.
//   - s (Session): Parameter.
//
// Returns:
//   - context.Context: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func NewContextWithSession(ctx context.Context, s Session) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, s)
}

// GetSession retrieves the Session from the context.
//
// Summary: Retrieves the Session from the context.
//
// Parameters:
//   - ctx (context.Context): Parameter.
//
// Returns:
//   - Session: Return value.
//   - bool: Return value.
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

// NewContextWithSampler creates a new context with the given Sampler.
//
// Summary: Creates a new context with the given Sampler.
//
// Parameters:
//   - ctx (context.Context): Parameter.
//   - s (Sampler): Parameter.
//
// Returns:
//   - context.Context: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func NewContextWithSampler(ctx context.Context, s Sampler) context.Context {
	return NewContextWithSession(ctx, s)
}

// GetSampler retrieves the Sampler from the context.
//
// Summary: Retrieves the Sampler from the context.
//
// Parameters:
//   - ctx (context.Context): Parameter.
//
// Returns:
//   - Sampler: Return value.
//   - bool: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func GetSampler(ctx context.Context) (Sampler, bool) {
	return GetSession(ctx)
}
