// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"context"

	"google.golang.org/grpc"
)

// WrappedServerStream is a wrapper around grpc.ServerStream that allows modifying the context.
//
// Summary: A gRPC server stream wrapper supporting context modification.
//
// Description:
// This struct embeds `grpc.ServerStream` and overrides the `Context()` method
// to return a custom context. This is useful for middleware that needs to inject
// values into the stream context.
type WrappedServerStream struct {
	grpc.ServerStream
	Ctx context.Context
}

// Context returns the modified context.
//
// Summary: Retrieves the stream context.
//
// Returns:
//   - context.Context: The modified context.
func (w *WrappedServerStream) Context() context.Context {
	return w.Ctx
}
