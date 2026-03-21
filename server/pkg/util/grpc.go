// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"context"

	"google.golang.org/grpc"
)

// Summary: WrappedServerStream is a wrapper around grpc.ServerStream that allows modifying the context. A wrapper for grpc.ServerStream that overrides the context.
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
type WrappedServerStream struct {
	grpc.ServerStream
	Ctx context.Context
}

// Summary: Context returns the modified context. Returns the context associated with the stream.
//
// Parameters:
//   - None.
//
// Returns:
//   - context.Context: The resulting context.Context.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (w *WrappedServerStream) Context() context.Context {
	return w.Ctx
}
