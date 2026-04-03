// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"context"

	"google.golang.org/grpc"
)

// WrappedServerStream represents the public WrappedServerStream entity.
//
// Summary: Defines the structured data model representing a server stream.
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

// Context serves as a public interface for interacting with Context.
//
// Summary: Context the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (w *WrappedServerStream) Context() context.Context {
	return w.Ctx
}
