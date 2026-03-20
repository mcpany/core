// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestActiveIntentAlignmentMiddleware_Pass(t *testing.T) {
	middleware := NewActiveIntentAlignmentMiddleware("super_secret_mission_root")

	// Valid heartbeat
	argsJSON := `{"target":"db","x-intent-heartbeat":"valid_signature","x-agent-id":"agent-1","x-intent-context":"fetch_user"}`
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "query_db",
			Arguments: json.RawMessage(argsJSON),
		},
	}

	nextCalled := false
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return &mcp.CallToolResult{}, nil
	}

	res, err := middleware.Execute(context.Background(), "tools/call", req, next)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.True(t, nextCalled)
}

func TestActiveIntentAlignmentMiddleware_MissingHeartbeat(t *testing.T) {
	middleware := NewActiveIntentAlignmentMiddleware("super_secret_mission_root")

	// Missing heartbeat entirely
	argsJSON := `{"target":"db"}`
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "query_db",
			Arguments: json.RawMessage(argsJSON),
		},
	}

	nextCalled := false
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return &mcp.CallToolResult{}, nil
	}

	res, err := middleware.Execute(context.Background(), "tools/call", req, next)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.False(t, nextCalled)
	assert.Contains(t, err.Error(), "Intent Drift Detected: Action is missing")
}

func TestActiveIntentAlignmentMiddleware_InvalidSignature(t *testing.T) {
	middleware := NewActiveIntentAlignmentMiddleware("super_secret_mission_root")

	// Invalid heartbeat signature
	argsJSON := `{"target":"db","x-intent-heartbeat":"INVALID_HEARTBEAT_SIGNATURE"}`
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "query_db",
			Arguments: json.RawMessage(argsJSON),
		},
	}

	nextCalled := false
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return &mcp.CallToolResult{}, nil
	}

	res, err := middleware.Execute(context.Background(), "tools/call", req, next)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.False(t, nextCalled)
	assert.Contains(t, err.Error(), "Heartbeat signature mismatch")
}

func TestActiveIntentAlignmentMiddleware_NotToolsCall(t *testing.T) {
	middleware := NewActiveIntentAlignmentMiddleware("super_secret_mission_root")

	req := &mcp.ListToolsRequest{}

	nextCalled := false
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return &mcp.ListToolsResult{}, nil
	}

	res, err := middleware.Execute(context.Background(), "tools/list", req, next)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.True(t, nextCalled)
}
