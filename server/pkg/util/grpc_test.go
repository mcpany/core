// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util_test

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

// mockServerStream implements grpc.ServerStream for testing purposes.
type mockServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context {
	return m.ctx
}

func TestWrappedServerStream_Context(t *testing.T) {
	t.Parallel()
	baseCtx := context.WithValue(context.Background(), "base", "value")
	newCtx := context.WithValue(context.Background(), "new", "value")

	mockStream := &mockServerStream{ctx: baseCtx}
	wrappedStream := &util.WrappedServerStream{
		ServerStream: mockStream,
		Ctx:          newCtx,
	}

	// The WrappedServerStream should return the new context, not the base context from the embedded stream
	assert.Equal(t, newCtx, wrappedStream.Context())
	assert.NotEqual(t, baseCtx, wrappedStream.Context())

	// Verify that the underlying stream is preserved
	// In a real scenario we might call other methods, but here we just checking the structure
	assert.Equal(t, mockStream, wrappedStream.ServerStream)
}
