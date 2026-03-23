// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
)

func TestLazyMCPMiddleware_Enabled(t *testing.T) {
	config := LazyMCPConfig{
		Enabled:   true,
		Threshold: 0.85,
		CacheTTL:  "600s",
	}

	middleware := NewLazyMCPMiddleware(config)

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
