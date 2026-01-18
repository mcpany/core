// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/util"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestToolManager_ExecuteTool_FuzzyMatch(t *testing.T) {
	t.Parallel()
	tm := NewManager(nil)

	sanitizedToolName, _ := util.SanitizeToolName("get_weather")
	serviceID := "weather"
	toolID := serviceID + "." + sanitizedToolName

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

	assert.Contains(t, err.Error(), "unknown tool")
	assert.Contains(t, err.Error(), "did you mean")
	assert.Contains(t, err.Error(), toolID)
}
