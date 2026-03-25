// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"context"
// WrappedServerStream is a wrapper around grpc.ServerStream that allows modifying the context.
// Context returns the modified context.
//
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Summary: Returns the context associated with the stream.
//
// Returns:
//   - context.Context: The modified context.
// Parameters:
//   - standard arguments based on function signature.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (w *WrappedServerStream) Context() context.Context {
	return w.Ctx
}
