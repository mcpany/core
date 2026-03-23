// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
)

func TestGranularScopesMiddleware_Allowed(t *testing.T) {
	config := GranularScopesConfig{
		Default: "deny",
		Tokens:  []string{"fs:read:/tmp"},
	}

	middleware := NewGranularScopesMiddleware(config)

	ctx := context.Background()
	req := &tool.ExecutionRequest{
		ToolName: "fs.read_file",
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

func TestGranularScopesMiddleware_Denied(t *testing.T) {
	config := GranularScopesConfig{
		Default: "deny",
		Tokens:  []string{"fs:read:/tmp"},
	}

	middleware := NewGranularScopesMiddleware(config)

	ctx := context.Background()
	req := &tool.ExecutionRequest{
		ToolName: "db.write_user",
	}

	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "success", nil
	}

	res, err := middleware.Execute(ctx, req, next)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "execution denied by granular scopes")
	assert.Nil(t, res)
}
