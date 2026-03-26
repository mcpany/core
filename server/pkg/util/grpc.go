// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"context"

	"google.golang.org/grpc"
)

// WrappedServerStream is a wrapper around grpc.ServerStream that allows modifying the context.
// Summary: A wrapper for grpc.ServerStream that overrides the context.
// WrappedServerStream is a wrapper around grpc.ServerStream that allows modifying the context.
// Summary: A wrapper for grpc.ServerStream that overrides the context.
	grpc.ServerStream
	Ctx context.Context
}

// Context returns the modified context.
// Summary: Returns the context associated with the stream.
// Returns:
//   - context.Context: The modified context.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// Context returns the modified context.
// Summary: Returns the context associated with the stream.
// Returns:
//   - context.Context: The modified context.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// Parameters:
//   - None.
	return w.Ctx
}
