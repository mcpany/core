// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"context"

	"google.golang.org/grpc"
)

// WrappedServerStream is a wrapper around grpc.ServerStream that allows modifying the context.
//
// Summary: is a wrapper around grpc.ServerStream that allows modifying the context.
type WrappedServerStream struct {
	grpc.ServerStream
	Ctx context.Context
}

// Context returns the modified context.
//
// Summary: returns the modified context.
//
// Parameters:
//   None.
//
// Returns:
//   - context.Context: The context.Context.
func (w *WrappedServerStream) Context() context.Context {
	return w.Ctx
}
