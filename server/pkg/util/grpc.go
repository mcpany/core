// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"context"

	"google.golang.org/grpc"
)

// WrappedServerStream is a wrapper around grpc.ServerStream that allows modifying the context.
//
// Summary: Is a wrapper around grpc.ServerStream that allows modifying the context.
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

// Context returns the modified context.
//
// Summary: Returns the modified context.
//
// Parameters:
//   - None.
//
// Returns:
//   - context.Context: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (w *WrappedServerStream) Context() context.Context {
	return w.Ctx
}
