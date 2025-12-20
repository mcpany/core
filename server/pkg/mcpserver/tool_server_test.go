// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/tool"
	v1 "github.com/mcpany/core/proto/api/v1"
	mcprouterv1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
)

func TestToolServer_ListTools(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)
	server := NewToolServer(mockManager)

	mockTool := &tool.MockTool{
		ToolFunc: func() *mcprouterv1.Tool {
			return &mcprouterv1.Tool{Name: proto.String("test-tool")}
		},
	}

	mockManager.EXPECT().ListTools().Return([]tool.Tool{mockTool})

	resp, err := server.ListTools(context.Background(), &v1.ListToolsRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.Tools, 1)
	assert.Equal(t, "test-tool", resp.Tools[0].GetName())
}

func TestToolServer_GetTool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)
	server := NewToolServer(mockManager)

	mockTool := &tool.MockTool{
		ToolFunc: func() *mcprouterv1.Tool {
			return &mcprouterv1.Tool{Name: proto.String("test-tool")}
		},
	}

	mockManager.EXPECT().GetTool("test-tool").Return(mockTool, true)

	resp, err := server.GetTool(context.Background(), &v1.GetToolRequest{ToolName: "test-tool"})
	require.NoError(t, err)
	assert.Equal(t, "test-tool", resp.Tool.GetName())
}

func TestToolServer_GetTool_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)
	server := NewToolServer(mockManager)

	mockManager.EXPECT().GetTool("unknown").Return(nil, false)

	_, err := server.GetTool(context.Background(), &v1.GetToolRequest{ToolName: "unknown"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestToolServer_RegisterTools_Unimplemented(t *testing.T) {
	server := NewToolServer(nil)
	_, err := server.RegisterTools(context.Background(), &v1.RegisterToolsRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}
