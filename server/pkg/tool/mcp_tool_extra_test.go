// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/server/pkg/client"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

type mockMCPClient2 struct {
	client.MCPClient
	result *mcp.CallToolResult
}

func (m *mockMCPClient2) CallTool(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	return m.result, nil
}

func TestMCPTool_OutputTransformation(t *testing.T) {
	mockClient := &mockMCPClient2{
		result: &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: `{"raw":"data"}`},
			},
		},
	}

	toolProto := &v1.Tool{Name: proto.String("test")}
	toolProto.SetServiceId("s1")

	format := configv1.OutputTransformer_JSON
	callDef := &configv1.MCPCallDefinition{
		OutputTransformer: &configv1.OutputTransformer{
			Format: &format,
			ExtractionRules: map[string]string{
				"extracted": "{.raw}",
			},
		},
	}

	mcpTool := NewMCPTool(toolProto, mockClient, callDef)

	req := &ExecutionRequest{ToolInputs: json.RawMessage(`{}`)}
	res, err := mcpTool.Execute(context.Background(), req)
	assert.NoError(t, err)

	assert.Equal(t, map[string]any{"extracted": "data"}, res)
}

func TestMCPTool_OutputTransformation_RawBytes(t *testing.T) {
	mockClient := &mockMCPClient2{
		result: &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: `some raw text`},
			},
		},
	}

	format := configv1.OutputTransformer_RAW_BYTES
	callDef := &configv1.MCPCallDefinition{
		OutputTransformer: &configv1.OutputTransformer{
			Format: &format,
		},
	}

	mcpTool := NewMCPTool(&v1.Tool{Name: proto.String("test")}, mockClient, callDef)

	req := &ExecutionRequest{ToolInputs: json.RawMessage(`{}`)}
	res, err := mcpTool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resMap := res.(map[string]any)
	assert.Equal(t, []byte("some raw text"), resMap["raw"])
}
