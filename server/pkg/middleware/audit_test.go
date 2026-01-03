// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/middleware"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditMiddleware(t *testing.T) {
	// Create temp file for logs
	tmpFile, err := os.CreateTemp("", "audit_test_*.log")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close() // AuditMiddleware will open it

	enabled := true
	logArgs := true
	logRes := true
	path := tmpFile.Name()

	cfg := &configv1.AuditConfig{
		Enabled:      &enabled,
		OutputPath:   &path,
		LogArguments: &logArgs,
		LogResults:   &logRes,
	}

	m, err := middleware.NewAuditMiddleware(cfg)
	require.NoError(t, err)
	defer func() { _ = m.Close() }()

	ctx := context.Background()
	ctx = auth.ContextWithUser(ctx, "user1")
	ctx = auth.ContextWithProfileID(ctx, "profileA")

	inputs := map[string]any{"arg1": "val1"}
	inputsBytes, _ := json.Marshal(inputs)

	req := &tool.ExecutionRequest{
		ToolName:   "test_tool",
		ToolInputs: json.RawMessage(inputsBytes),
	}

	mockNext := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return map[string]any{"status": "ok"}, nil
	}

	result, err := m.Execute(ctx, req, mockNext)
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"status": "ok"}, result)

	// Verify log file content
	content, err := os.ReadFile(tmpFile.Name())
	require.NoError(t, err)

	var entry middleware.AuditEntry
	err = json.Unmarshal(content, &entry)
	require.NoError(t, err)

	assert.Equal(t, "test_tool", entry.ToolName)
	assert.Equal(t, "user1", entry.UserID)
	assert.Equal(t, "profileA", entry.ProfileID)
	assert.NotEmpty(t, entry.Duration)

	// Check args
	var args map[string]any
	err = json.Unmarshal(entry.Arguments, &args)
	require.NoError(t, err)
	assert.Equal(t, "val1", args["arg1"])

	// Check result
	resMap, ok := entry.Result.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "ok", resMap["status"])
}

func TestAuditMiddleware_Disabled(t *testing.T) {
	enabled := false
	cfg := &configv1.AuditConfig{
		Enabled: &enabled,
	}
	m, err := middleware.NewAuditMiddleware(cfg)
	require.NoError(t, err)

	// mockNext called
	called := false
	mockNext := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		called = true
		return nil, nil
	}

	_, _ = m.Execute(context.Background(), nil, mockNext)
	assert.True(t, called)
}
