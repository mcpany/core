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
// Summary. Represents a Session.
type Session interface {
	// CreateMessage requests a message creation (sampling) from the client.
	//
	// Summary: Requests message creation.
	//
// Parameters.
	//   - ctx: context.Context. The context for the request.
	//   - params: *mcp.CreateMessageParams. The parameters for message creation.
	//
// Returns.
	//   - *mcp.CreateMessageResult: The result of the message creation.
	//   - error: An error if the operation fails.
	CreateMessage(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error)

	// ListRoots requests the list of roots from the client.
	//
	// Summary: Requests roots list.
	//
// Parameters.
	//   - ctx: context.Context. The context for the request.
	//
// Returns.
	//   - *mcp.ListRootsResult: The list of roots.
	//   - error: An error if the operation fails.
	ListRoots(ctx context.Context) (*mcp.ListRootsResult, error)
}

// Sampler is an alias for Session for backward compatibility.
//
// Summary: Represents a Sampler.
type Sampler = Session

type sessionContextKey struct{}

// NewContextWithSession provides newcontextwithsession functionality.
//
// Summary: NewContextWithSession.
//
// Parameters.
//   - ctx: The parameter.
//   - s: The parameter.
//
// Returns.
//   - result: The result.
func NewContextWithSession(ctx context.Context, s Session) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, s)
}

// GetSession provides getsession functionality.
//
// Summary: GetSession.
//
// Parameters.
//   - ctx: The parameter.
//   - bool: The parameter.
//
// Returns.
//   - None.
func GetSession(ctx context.Context) (Session, bool) {
	s, ok := ctx.Value(sessionContextKey{}).(Session)
	return s, ok
}

// NewContextWithSampler provides newcontextwithsampler functionality.
//
// Summary: NewContextWithSampler.
//
// Parameters.
//   - ctx: The parameter.
//   - s: The parameter.
//
// Returns.
//   - result: The result.
func NewContextWithSampler(ctx context.Context, s Sampler) context.Context {
	return NewContextWithSession(ctx, s)
}

// GetSampler provides getsampler functionality.
//
// Summary: GetSampler.
//
// Parameters.
//   - ctx: The parameter.
//   - bool: The parameter.
//
// Returns.
//   - None.
func GetSampler(ctx context.Context) (Sampler, bool) {
	return GetSession(ctx)
}
