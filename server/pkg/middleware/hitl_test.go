// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"testing"
	"time"

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

	// Set up a background worker to simulate the human approving the request
	go func() {
		// Wait a tiny bit to ensure the middleware has subscribed
		time.Sleep(2000 * time.Millisecond)

		reqBus, _ := corebus.GetBus[HITLApprovalRequest](bp, "hitl.requests")
		// Listen for the request to get the execution ID
		reqBus.SubscribeOnce(context.Background(), "hitl.requests", func(req HITLApprovalRequest) {
			resBus, _ := corebus.GetBus[HITLApprovalResponse](bp, "hitl.responses."+req.ExecutionID)
			// Simulate approval
			_ = resBus.Publish(context.Background(), "hitl.responses."+req.ExecutionID, HITLApprovalResponse{
				ExecutionID: req.ExecutionID,
				Approved:    true,
			})
		})
	}()

	mockNext := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "success", nil
	}

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

	// Set up a background worker to simulate the human denying the request
	go func() {
		time.Sleep(2000 * time.Millisecond)
		reqBus, _ := corebus.GetBus[HITLApprovalRequest](bp, "hitl.requests")
		reqBus.SubscribeOnce(context.Background(), "hitl.requests", func(req HITLApprovalRequest) {
			resBus, _ := corebus.GetBus[HITLApprovalResponse](bp, "hitl.responses."+req.ExecutionID)
			_ = resBus.Publish(context.Background(), "hitl.responses."+req.ExecutionID, HITLApprovalResponse{
				ExecutionID: req.ExecutionID,
				Approved:    false, // Denied
			})
		})
	}()

	mockNext := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "success", nil
	}

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

	// We intentionally do NOT set up a background worker to approve, so it will timeout.

	mockNext := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "success", nil
	}

	res, err := middleware.Execute(ctx, req, mockNext)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "timeout reached or context cancelled")
}
