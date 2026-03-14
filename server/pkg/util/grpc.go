// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util	//nolint:revive,nolintlint // Package name 'util' is common in this codebase
// WrappedServerStream is a wrapper around grpc.ServerStream that allows modifying the context.
//
// Summary: A wrapper for grpc.ServerStream that overrides the context.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Context returns the modified context.
//
// Summary: Returns the context associated with the stream.
//
// Returns:
//   - context.Context: The modified context.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
import (
	"context"

	"google.golang.org/grpc"
)

type WrappedServerStream struct {
	grpc.ServerStream
	Ctx	context.Context
}

func (w *WrappedServerStream) Context() context.Context {
	return w.Ctx
}
