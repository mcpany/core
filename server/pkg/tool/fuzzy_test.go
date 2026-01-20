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

func TestToolManager_ExecuteTool_NamespaceSuggestion(t *testing.T) {
	t.Parallel()
	tm := NewManager(nil)

	serviceID := "weather"
	toolName := "get_current"

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

	// Scenario 1: User forgets namespace
	// We expect the error to suggest "weather.get_current"
	execReq := &ExecutionRequest{ToolName: toolName, ToolInputs: []byte(`{}`)}

	_, err = tm.ExecuteTool(context.Background(), execReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `did you mean "weather.get_current"?`)
}

func TestToolManager_ExecuteTool_MultipleSuggestions(t *testing.T) {
	t.Parallel()
	tm := NewManager(nil)

	// Register multiple similar tools
	tools := []struct{ svc, name string }{
		{"git", "status"},
		{"git", "stash"},
		{"got", "stat"}, // unrelated service, similar name
	}

	for _, tc := range tools {
		mockTool := &MockTool{
			ToolFunc: func() *v1.Tool {
				return &v1.Tool{
					ServiceId: proto.String(tc.svc),
					Name:      proto.String(tc.name),
				}
			},
			ExecuteFunc: func(_ context.Context, _ *ExecutionRequest) (any, error) { return "ok", nil },
		}
		err := tm.AddTool(mockTool)
		assert.NoError(t, err)
	}

	// Scenario 2: User types "git.stat"
	// Should suggest "git.status" and maybe "git.stash" (dist 2)
	// "got.stat" is dist 2.
	execReq := &ExecutionRequest{ToolName: "git.stat", ToolInputs: []byte(`{}`)}
	_, err := tm.ExecuteTool(context.Background(), execReq)
	assert.Error(t, err)

	// We verify that AT LEAST "git.status" is suggested
	assert.Contains(t, err.Error(), "git.status")
}
