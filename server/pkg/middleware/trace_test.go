// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithTraceContext(t *testing.T) {
	t.Parallel()

	t.Run("with_parent_id", func(t *testing.T) {
		ctx := WithTraceContext(context.Background(), "trace-123", "span-456", "parent-789")
		assert.Equal(t, "trace-123", GetTraceID(ctx))
		assert.Equal(t, "span-456", GetSpanID(ctx))
		assert.Equal(t, "parent-789", GetParentID(ctx))
	})

	t.Run("without_parent_id", func(t *testing.T) {
		ctx := WithTraceContext(context.Background(), "trace-abc", "span-def", "")
		assert.Equal(t, "trace-abc", GetTraceID(ctx))
		assert.Equal(t, "span-def", GetSpanID(ctx))
		assert.Empty(t, GetParentID(ctx))
	})
}

func TestTraceContext_GetEmpty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	assert.Empty(t, GetTraceID(ctx))
	assert.Empty(t, GetSpanID(ctx))
	assert.Empty(t, GetParentID(ctx))
}

func TestTraceContext_WrongTypes(t *testing.T) {
	t.Parallel()

	// Setting keys manually to test the type assertion failure branch
	ctx := context.WithValue(context.Background(), traceIDKey, 123)
	ctx = context.WithValue(ctx, spanIDKey, 456)
	ctx = context.WithValue(ctx, parentIDKey, 789)

	assert.Empty(t, GetTraceID(ctx))
	assert.Empty(t, GetSpanID(ctx))
	assert.Empty(t, GetParentID(ctx))
}
