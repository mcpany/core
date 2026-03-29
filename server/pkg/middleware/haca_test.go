// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
)

func TestHACAMiddleware_Enabled_Success(t *testing.T) {
	config := HACAConfig{
		Enabled:          true,
		MaxTokens:        1000,
		MaxReasoningTime: 500,
	}
	middleware := NewHACAMiddleware(config)

	req := &tool.ExecutionRequest{
		ToolName: "test-tool",
		Arguments: map[string]interface{}{
			"_x_mission_lineage":         "branch-A-123",
			"_x_gemini_reasoning_effort": 100,
			"_x_tokens_consumed":         200,
			"target":                     "something",
		},
	}

	executed := false
	next := func(ctx context.Context, r *tool.ExecutionRequest) (any, error) {
		executed = true
		assert.NotContains(t, r.Arguments, "_x_mission_lineage")
		assert.NotContains(t, r.Arguments, "_x_gemini_reasoning_effort")
		assert.NotContains(t, r.Arguments, "_x_tokens_consumed")
		assert.Contains(t, r.Arguments, "target")
		return "success", nil
	}

	res, err := middleware.Execute(context.Background(), req, next)

	assert.NoError(t, err)
	assert.Equal(t, "success", res)
	assert.True(t, executed)
}

func TestHACAMiddleware_TokenExhaustion(t *testing.T) {
	config := HACAConfig{
		Enabled:          true,
		MaxTokens:        100,
	}
	middleware := NewHACAMiddleware(config)

	req := &tool.ExecutionRequest{
		ToolName: "test-tool",
		Arguments: map[string]interface{}{
			"_x_mission_lineage": "branch-A-123",
			"_x_tokens_consumed": 150, // exceeds 100
		},
	}

	next := func(ctx context.Context, r *tool.ExecutionRequest) (any, error) {
		return "success", nil
	}

	res, err := middleware.Execute(context.Background(), req, next)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Token budget exhausted")
	assert.Nil(t, res)
}

func TestHACAMiddleware_ReasoningTimeExhaustion(t *testing.T) {
	config := HACAConfig{
		Enabled:          true,
		MaxReasoningTime: 50,
	}
	middleware := NewHACAMiddleware(config)

	req := &tool.ExecutionRequest{
		ToolName: "test-tool",
		Arguments: map[string]interface{}{
			"_x_mission_lineage":         "branch-B-456",
			"_x_gemini_reasoning_effort": 60, // exceeds 50
		},
	}

	next := func(ctx context.Context, r *tool.ExecutionRequest) (any, error) {
		return "success", nil
	}

	res, err := middleware.Execute(context.Background(), req, next)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Reasoning time budget exhausted")
	assert.Nil(t, res)
}

func TestHACAMiddleware_MissingLineage(t *testing.T) {
	config := HACAConfig{
		Enabled: true,
	}
	middleware := NewHACAMiddleware(config)

	req := &tool.ExecutionRequest{
		ToolName: "test-tool",
		Arguments: map[string]interface{}{
			"target": "something",
		},
	}

	next := func(ctx context.Context, r *tool.ExecutionRequest) (any, error) {
		return "success", nil
	}

	res, err := middleware.Execute(context.Background(), req, next)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing mission lineage")
	assert.Nil(t, res)
}
