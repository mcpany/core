// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"context"

	"google.golang.org/grpc"
)

// WrappedServerStream wrappedServerStream represents a wrapped server stream.
//
// Summary: WrappedServerStream represents a wrapped server stream.
type WrappedServerStream struct {
	grpc.ServerStream
	Ctx context.Context
}

// Context context context.
//
// Summary: Context context.
//
// Parameters:
//   - None.
//
// Returns:
//   - context.Context: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (w *WrappedServerStream) Context() context.Context {
	return w.Ctx
}
