// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"context"

	"google.golang.org/grpc"
)

// WrappedServerStream - Auto-generated documentation.
//
// Summary: WrappedServerStream is a wrapper around grpc.ServerStream that allows modifying the context.
//
// Fields:
//   - Various fields for WrappedServerStream.
type WrappedServerStream struct {
	grpc.ServerStream
	Ctx context.Context
}

// Context - Auto-generated documentation.
//
// Summary: Context returns the modified context.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (w *WrappedServerStream) Context() context.Context {
	return w.Ctx
}
