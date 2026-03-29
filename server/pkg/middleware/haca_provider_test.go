// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"
	"testing"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
)

func TestHACAProvider_Disabled(t *testing.T) {
	provider := NewHACAProviderMiddleware(HACAProviderConfig{Enabled: false})

	req := &tool.ExecutionRequest{
		ToolName: "test_tool",
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return "success", nil
	}

	res, err := provider.Execute(context.Background(), req, next)

	assert.NoError(t, err)
	assert.Equal(t, "success", res)
	assert.True(t, nextCalled)
}

func TestHACAProvider_MissingArguments(t *testing.T) {
	provider := NewHACAProviderMiddleware(HACAProviderConfig{Enabled: true})

	req := &tool.ExecutionRequest{
		ToolName:  "test_tool",
		Arguments: nil,
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return nil, nil
	}

	res, err := provider.Execute(context.Background(), req, next)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "Economic Security Violation: missing hardware-attested cost token")
	assert.False(t, nextCalled)
}

func TestHACAProvider_MissingTokens(t *testing.T) {
	provider := NewHACAProviderMiddleware(HACAProviderConfig{Enabled: true})

	req := &tool.ExecutionRequest{
		ToolName: "test_tool",
		Arguments: map[string]interface{}{
			"param1": "value1",
		},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return nil, nil
	}

	res, err := provider.Execute(context.Background(), req, next)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "Economic Security Violation: missing cost attribution token or lineage ID")
	assert.False(t, nextCalled)
}

func TestHACAProvider_InvalidTokens(t *testing.T) {
	provider := NewHACAProviderMiddleware(HACAProviderConfig{Enabled: true})

	req := &tool.ExecutionRequest{
		ToolName: "test_tool",
		Arguments: map[string]interface{}{
			"_x_haca_token": 123,
			"_x_lineage_id": "lineage1",
		},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return nil, nil
	}

	res, err := provider.Execute(context.Background(), req, next)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "Economic Security Violation: invalid cost attribution token format")
	assert.False(t, nextCalled)
}

func TestHACAProvider_AttestationFailed(t *testing.T) {
	provider := NewHACAProviderMiddleware(HACAProviderConfig{Enabled: true})

	req := &tool.ExecutionRequest{
		ToolName: "test_tool",
		Arguments: map[string]interface{}{
			"_x_haca_token": "invalid-prefix-token",
			"_x_lineage_id": "lineage1",
		},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return nil, nil
	}

	res, err := provider.Execute(context.Background(), req, next)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), fmt.Sprintf("Economic Security Violation: token attestation failed for lineage '%s'", "lineage1"))
	assert.False(t, nextCalled)
}

func TestHACAProvider_Success(t *testing.T) {
	provider := NewHACAProviderMiddleware(HACAProviderConfig{Enabled: true})

	req := &tool.ExecutionRequest{
		ToolName: "test_tool",
		Arguments: map[string]interface{}{
			"_x_haca_token": "haca-attested-token-123",
			"_x_lineage_id": "lineage-root-1",
			"param1":        "value1",
		},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true

		// Verify args were cleaned
		args := req.Arguments
		_, hasToken := args["_x_haca_token"]
		_, hasLineage := args["_x_lineage_id"]
		assert.False(t, hasToken)
		assert.False(t, hasLineage)
		assert.Equal(t, "value1", args["param1"])

		return "success", nil
	}

	res, err := provider.Execute(context.Background(), req, next)

	assert.NoError(t, err)
	assert.Equal(t, "success", res)
	assert.True(t, nextCalled)
}

func TestHACAProvider_RequiredFor_Skipped(t *testing.T) {
	provider := NewHACAProviderMiddleware(HACAProviderConfig{
		Enabled:     true,
		RequiredFor: []string{"critical_tool"},
	})

	req := &tool.ExecutionRequest{
		ToolName: "safe_tool",
		Arguments: map[string]interface{}{},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return "success", nil
	}

	res, err := provider.Execute(context.Background(), req, next)

	assert.NoError(t, err)
	assert.Equal(t, "success", res)
	assert.True(t, nextCalled)
}

func TestHACAProvider_RequiredFor_Enforced(t *testing.T) {
	provider := NewHACAProviderMiddleware(HACAProviderConfig{
		Enabled:     true,
		RequiredFor: []string{"critical_tool"},
	})

	req := &tool.ExecutionRequest{
		ToolName: "critical_tool",
		Arguments: map[string]interface{}{},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return nil, nil
	}

	res, err := provider.Execute(context.Background(), req, next)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "Economic Security Violation: missing cost attribution token or lineage ID")
	assert.False(t, nextCalled)
}
