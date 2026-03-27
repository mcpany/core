// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"context"

	"google.golang.org/grpc"
)

// WrappedServerStream is a wrapper around grpc.ServerStream that allows modifying the context.
//
// Summary: A wrapper for grpc.ServerStream that overrides the context.
type WrappedServerStream struct {
	grpc.ServerStream
	Ctx context.Context
}

// Context provides context functionality.
//
// Summary: Context.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (w *WrappedServerStream) Context() context.Context {
	return w.Ctx
}
