// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
)

func TestSharedKVStoreMiddleware_Disabled(t *testing.T) {
	config := SharedKVStoreConfig{
		Enabled: false,
	}

	middleware := NewSharedKVStoreMiddleware(config)

	ctx := context.Background()
	req := &tool.ExecutionRequest{
		ToolName: "test_tool",
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return "success", nil
	}

	res, err := middleware.Execute(ctx, req, next)
	assert.NoError(t, err)
	assert.Equal(t, "success", res)
	assert.True(t, nextCalled)
}

func TestSharedKVStoreMiddleware_AgentAware(t *testing.T) {
	config := SharedKVStoreConfig{
		Enabled:        true,
		IsolationLevel: "agent_aware",
	}

	middleware := NewSharedKVStoreMiddleware(config)

	ctx := context.Background()
	req := &tool.ExecutionRequest{
		ToolName: "database.query",
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return "success", nil
	}

	res, err := middleware.Execute(ctx, req, next)
	assert.NoError(t, err)
	assert.Equal(t, "success", res)
	assert.True(t, nextCalled)
}
