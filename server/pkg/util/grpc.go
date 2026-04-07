// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"context"

	"google.golang.org/grpc"
)

// Summary: WrappedServerStream represents a data structure.
//
// Parameters:
//   - None
//
// Returns:
//   - None
//
// Errors:
//   - None
//
// Side Effects:
//   - None
type WrappedServerStream struct {
	grpc.ServerStream
	Ctx context.Context
}

// Context returns the modified context.
//
// Summary: Context executes the operation.
//
// Parameters:
//   - None
//
// Returns:
//   - context.Context {
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (w *WrappedServerStream) Context() context.Context {
	return w.Ctx
}
