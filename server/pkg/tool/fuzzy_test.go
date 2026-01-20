// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
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
			return &v1.Tool{
				ServiceId: proto.String(serviceID),
				Name:      proto.String("get_weather"),
			}
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

func TestToolManager_ExecuteTool_SmartFuzzyMatch(t *testing.T) {
	t.Parallel()
	tm := NewManager(nil)

	serviceID := "weather"
	toolName := "get_forecast"

	mockTool := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				ServiceId: proto.String(serviceID),
				Name:      proto.String(toolName),
			}
		},
		ExecuteFunc: func(_ context.Context, _ *ExecutionRequest) (any, error) {
			return "sunny", nil
		},
	}

	err := tm.AddTool(mockTool)
	assert.NoError(t, err)

	// Case 1: Missing Namespace
	// User types "get_forecast" -> "weather.get_forecast"
	t.Run("Missing Namespace", func(t *testing.T) {
		execReq := &ExecutionRequest{ToolName: "get_forecast", ToolInputs: []byte(`{}`)}
		_, err := tm.ExecuteTool(context.Background(), execReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `did you mean "weather.get_forecast"?`)
	})

	// Case 2: Typo in Short Name
	// User types "get_forcast" -> "weather.get_forecast"
	t.Run("Typo in Short Name", func(t *testing.T) {
		execReq := &ExecutionRequest{ToolName: "get_forcast", ToolInputs: []byte(`{}`)}
		_, err := tm.ExecuteTool(context.Background(), execReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `did you mean "weather.get_forecast"?`)
	})

	// Case 3: Case Insensitive Short Name
	// User types "GetForecast" -> "weather.get_forecast"
	t.Run("Case Insensitive Short Name", func(t *testing.T) {
		execReq := &ExecutionRequest{ToolName: "GetForecast", ToolInputs: []byte(`{}`)}
		_, err := tm.ExecuteTool(context.Background(), execReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `did you mean "weather.get_forecast"?`)
	})

	// Case 4: Case Insensitive Full Name (if they typed namespace but wrong case)
	// User types "Weather.GetForecast" -> "weather.get_forecast"
	t.Run("Case Insensitive Full Name", func(t *testing.T) {
		execReq := &ExecutionRequest{ToolName: "Weather.GetForecast", ToolInputs: []byte(`{}`)}
		_, err := tm.ExecuteTool(context.Background(), execReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `did you mean "weather.get_forecast"?`)
	})

	// Case 5: Normalized Match (Hyphens/Underscores/Case)
	// User types "weather-get-forecast" -> "weather.get_forecast"
	t.Run("Normalized Match", func(t *testing.T) {
		execReq := &ExecutionRequest{ToolName: "Weather-Get-Forecast", ToolInputs: []byte(`{}`)}
		_, err := tm.ExecuteTool(context.Background(), execReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `did you mean "weather.get_forecast"?`)
	})

	// Case 6: Typo in Full Name (Standard Levenshtein)
	// User types "weather.get_forcast" -> "weather.get_forecast"
	t.Run("Typo in Full Name", func(t *testing.T) {
		execReq := &ExecutionRequest{ToolName: "weather.get_forcast", ToolInputs: []byte(`{}`)}
		_, err := tm.ExecuteTool(context.Background(), execReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `did you mean "weather.get_forecast"?`)
	})
}

func TestToolManager_ExecuteTool_FuzzyMatch_NoNamespace(t *testing.T) {
	t.Parallel()
	tm := NewManager(nil)

	// Tool with no Service ID (simulating internal tool or legacy)
	mockTool := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				Name:      proto.String("internal_tool"),
				ServiceId: proto.String("internal"), // Needs a dummy service ID to pass validation
			}
		},
		ExecuteFunc: func(_ context.Context, _ *ExecutionRequest) (any, error) {
			return "ok", nil
		},
	}

	err := tm.AddTool(mockTool)
	assert.NoError(t, err)

	// Case 7: Typo in Non-Namespaced Tool
	// Note: With the fix, we assign a Service ID "internal", so the exposed name becomes "internal.internal_tool".
	// The fuzzy matcher should suggest "internal.internal_tool" even if we typed "intenal_tool".
	t.Run("Typo in Non-Namespaced Tool", func(t *testing.T) {
		execReq := &ExecutionRequest{ToolName: "intenal_tool", ToolInputs: []byte(`{}`)}
		_, err := tm.ExecuteTool(context.Background(), execReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `did you mean "internal.internal_tool"?`)
	})
}
