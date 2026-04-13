// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"testing"

	"github.com/mcpany/core/proto/bus"
	corebus "github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestBus(t *testing.T) *corebus.Provider {
	busConfig := &bus.MessageBus{}
	busConfig.SetInMemory(&bus.InMemoryBus{})
	p, err := corebus.NewProvider(busConfig)
	require.NoError(t, err)
	return p
}

func TestHITLMiddleware_Disabled(t *testing.T) {
	bp := setupTestBus(t)
	config := HITLConfig{
		Enabled:        false,
		SensitiveTools: []string{"database.drop_table"},
	}
	middleware := NewHITLMiddleware(config, bp)

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
	bp := setupTestBus(t)
	config := HITLConfig{
		Enabled:        true,
		SensitiveTools: []string{"database.drop_table"},
	}
	middleware := NewHITLMiddleware(config, bp)

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

func TestHITLMiddleware_ApprovalGranted(t *testing.T) {
	bp := setupTestBus(t)
	config := HITLConfig{
		Enabled:        true,
		SensitiveTools: []string{"database.drop_table"},
		TimeoutSeconds: 5,
	}
	middleware := NewHITLMiddleware(config, bp)

	ctx := context.Background()
	req := &tool.ExecutionRequest{
		ToolName: "database.drop_table",
	}

	mockNext := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "success", nil
	}

	reqBus, err := corebus.GetBus[HITLApprovalRequest](bp, "hitl.requests")
	require.NoError(t, err)

	unsubscribe := reqBus.Subscribe(ctx, "hitl.requests", func(msg HITLApprovalRequest) {
		resBus, err := corebus.GetBus[HITLApprovalResponse](bp, "hitl.responses."+msg.ExecutionID)
		require.NoError(t, err)
		err = resBus.Publish(ctx, "hitl.responses."+msg.ExecutionID, HITLApprovalResponse{
			ExecutionID: msg.ExecutionID,
			Approved:    true,
		})
		require.NoError(t, err)
	})
	defer unsubscribe()

	res, err := middleware.Execute(ctx, req, mockNext)
	assert.NoError(t, err)
	assert.Equal(t, "success", res)
}

func TestHITLMiddleware_ApprovalDenied(t *testing.T) {
	bp := setupTestBus(t)
	config := HITLConfig{
		Enabled:        true,
		SensitiveTools: []string{"aws.iam.*"},
		TimeoutSeconds: 5,
	}
	middleware := NewHITLMiddleware(config, bp)

	ctx := context.Background()
	req := &tool.ExecutionRequest{
		ToolName: "aws.iam.delete_user",
	}

	mockNext := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "success", nil
	}

	reqBus, err := corebus.GetBus[HITLApprovalRequest](bp, "hitl.requests")
	require.NoError(t, err)

	unsubscribe := reqBus.Subscribe(ctx, "hitl.requests", func(msg HITLApprovalRequest) {
		resBus, err := corebus.GetBus[HITLApprovalResponse](bp, "hitl.responses."+msg.ExecutionID)
		require.NoError(t, err)
		err = resBus.Publish(ctx, "hitl.responses."+msg.ExecutionID, HITLApprovalResponse{
			ExecutionID: msg.ExecutionID,
			Approved:    false,
		})
		require.NoError(t, err)
	})
	defer unsubscribe()

	res, err := middleware.Execute(ctx, req, mockNext)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "human denied request")
}

func TestHITLMiddleware_Timeout(t *testing.T) {
	bp := setupTestBus(t)
	config := HITLConfig{
		Enabled:        true,
		SensitiveTools: []string{"database.drop_table"},
		TimeoutSeconds: 1, // Short timeout
	}
	middleware := NewHITLMiddleware(config, bp)

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
	assert.Contains(t, err.Error(), "timeout reached or context cancelled")
}
