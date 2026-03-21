// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"strings"
	"testing"

	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestToolManager_ExecuteTool_FuzzyMatch(t *testing.T) {
	t.Parallel()
	tm := NewManager(nil)

	serviceID := "weather"

	mockTool := &MockTool{
		ToolFunc: func() *v1.Tool {
			return v1.Tool_builder{
				ServiceId: proto.String(serviceID),
				Name:      proto.String("get_weather"),
			}.Build()
		},
		ExecuteFunc: func(_ context.Context, _ *ExecutionRequest) (any, error) {
			return "sunny", nil
		},
	}

	err := tm.AddTool(mockTool)
	assert.NoError(t, err)

	// Try to execute a non-existent tool with a typo
	typoName := serviceID + ".get_wether" // Typo in tool name
	execReq := &ExecutionRequest{ToolName: typoName, ToolInputs: []byte(`{}`)}

	_, err = tm.ExecuteTool(context.Background(), execReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `did you mean "weather.get_weather"?`)
}

func TestToolManager_ExecuteTool_FuzzyMatch_NamespaceMissing(t *testing.T) {
	t.Parallel()
	tm := NewManager(nil)

	serviceID := "weather"

	mockTool := &MockTool{
		ToolFunc: func() *v1.Tool {
			return v1.Tool_builder{
				ServiceId: proto.String(serviceID),
				Name:      proto.String("get_forecast"),
			}.Build()
		},
		ExecuteFunc: func(_ context.Context, _ *ExecutionRequest) (any, error) {
			return "sunny", nil
		},
	}

	err := tm.AddTool(mockTool)
	assert.NoError(t, err)

	// Case 1: Missing Namespace
	// User types "get_forecast", expecting it to work.
	// Actual tool ID: "weather.get_forecast"
	missingNamespaceName := "get_forecast"
	execReq := &ExecutionRequest{ToolName: missingNamespaceName, ToolInputs: []byte(`{}`)}

	_, err = tm.ExecuteTool(context.Background(), execReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `did you mean "weather.get_forecast"?`, "Should suggest namespaced tool when namespace is missing")
}

func TestToolManager_ExecuteTool_FuzzyMatch_MultipleMatches(t *testing.T) {
	t.Parallel()
	tm := NewManager(nil)

	// Add tool 1: weather.get_info
	mockTool1 := &MockTool{
		ToolFunc: func() *v1.Tool {
			return v1.Tool_builder{
				ServiceId: proto.String("weather"),
				Name:      proto.String("get_info"),
			}.Build()
		},
		ExecuteFunc: func(_ context.Context, _ *ExecutionRequest) (any, error) { return nil, nil },
	}
	assert.NoError(t, tm.AddTool(mockTool1))

	// Add tool 2: stocks.get_info
	mockTool2 := &MockTool{
		ToolFunc: func() *v1.Tool {
			return v1.Tool_builder{
				ServiceId: proto.String("stocks"),
				Name:      proto.String("get_info"),
			}.Build()
		},
		ExecuteFunc: func(_ context.Context, _ *ExecutionRequest) (any, error) { return nil, nil },
	}
	assert.NoError(t, tm.AddTool(mockTool2))

	// User types "get_info"
	execReq := &ExecutionRequest{ToolName: "get_info", ToolInputs: []byte(`{}`)}
	_, err := tm.ExecuteTool(context.Background(), execReq)
	assert.Error(t, err)

	errMsg := err.Error()
	assert.Contains(t, errMsg, "did you mean one of:")
	assert.True(t, strings.Contains(errMsg, "weather.get_info") || strings.Contains(errMsg, "stocks.get_info"))
}
