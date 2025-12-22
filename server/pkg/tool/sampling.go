// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Sampler defines the interface for tools to request sampling from the client.
type Sampler interface {
	// CreateMessage requests a message creation (sampling) from the client.
	CreateMessage(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error)
}

type samplerContextKey struct{}

// NewContextWithSampler creates a new context with the given Sampler.
// ctx is the context.
// Returns the result.
func NewContextWithSampler(ctx context.Context, s Sampler) context.Context {
	return context.WithValue(ctx, samplerContextKey{}, s)
}

// GetSampler retrieves the Sampler from the context.
// ctx is the context.
// Returns the result, the result.
func GetSampler(ctx context.Context) (Sampler, bool) {
	s, ok := ctx.Value(samplerContextKey{}).(Sampler)
	return s, ok
}
