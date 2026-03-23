// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
)

func TestHITLMiddleware_Disabled(t *testing.T) {
	config := HITLConfig{
		Enabled:        false,
		SensitiveTools: []string{"database.drop_table"},
	}
	middleware := NewHITLMiddleware(config)

	ctx := context.Background()
	req := &tool.ExecutionRequest{
		ToolName: "database.drop_table",
	}

	mockNext := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "success", nil
	}

	res, err := middleware.Execute(ctx, req, mockNext)
	assert.NoError(t, err)
	assert.Equal(t, "success", res)
}

func TestHITLMiddleware_NotSensitive(t *testing.T) {
	config := HITLConfig{
		Enabled:        true,
		SensitiveTools: []string{"database.drop_table"},
	}
	middleware := NewHITLMiddleware(config)

	ctx := context.Background()
	req := &tool.ExecutionRequest{
		ToolName: "database.select",
	}

	mockNext := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "success", nil
	}

	res, err := middleware.Execute(ctx, req, mockNext)
	assert.NoError(t, err)
	assert.Equal(t, "success", res)
}

func TestHITLMiddleware_SensitiveExactMatch(t *testing.T) {
	config := HITLConfig{
		Enabled:        true,
		SensitiveTools: []string{"database.drop_table"},
		RequireMFA:     true,
	}
	middleware := NewHITLMiddleware(config)

	ctx := context.Background()
	req := &tool.ExecutionRequest{
		ToolName: "database.drop_table",
	}

	mockNext := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "success", nil
	}

	res, err := middleware.Execute(ctx, req, mockNext)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "execution suspended for HITL approval")
}

func TestHITLMiddleware_SensitivePrefixMatch(t *testing.T) {
	config := HITLConfig{
		Enabled:        true,
		SensitiveTools: []string{"aws.iam.*"},
	}
	middleware := NewHITLMiddleware(config)

	ctx := context.Background()
	req := &tool.ExecutionRequest{
		ToolName: "aws.iam.delete_user",
	}

	mockNext := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "success", nil
	}

	res, err := middleware.Execute(ctx, req, mockNext)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "execution suspended for HITL approval")
}
