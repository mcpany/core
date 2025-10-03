/*
 * Copyright 2025 Author(s) of MCPX
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tool_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/mcpxy/mcpx/pkg/client"
	"github.com/mcpxy/mcpx/pkg/tool"
	configv1 "github.com/mcpxy/mcpx/proto/config/v1"
	v1 "github.com/mcpxy/mcpx/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockMCPClient for testing
type mockMCPClient struct {
	client.MCPClient
	callToolFunc func(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error)
}

func (m *mockMCPClient) CallTool(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	if m.callToolFunc != nil {
		return m.callToolFunc(ctx, params)
	}
	return nil, errors.New("not implemented")
}

func TestMCPTool_Execute(t *testing.T) {
	t.Run("successful execution", func(t *testing.T) {
		mockClient := &mockMCPClient{
			callToolFunc: func(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
				assert.Equal(t, "test-tool", params.Name)
				args, ok := params.Arguments.(json.RawMessage)
				require.True(t, ok)
				assert.JSONEq(t, `{"input":"value"}`, string(args))
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: `{"output":"result"}`},
					},
				}, nil
			},
		}

		toolProto := &v1.Tool{}
		toolProto.SetName("test-tool")
		toolProto.SetServiceId("service")
		mcpTool := tool.NewMCPTool(toolProto, mockClient, &configv1.MCPCallDefinition{})

		inputs := json.RawMessage(`{"input":"value"}`)
		req := &tool.ExecutionRequest{
			ToolName:   "service/-/test-tool",
			ToolInputs: inputs,
		}

		result, err := mcpTool.Execute(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"output": "result"}, result)
	})

	t.Run("execution error", func(t *testing.T) {
		expectedErr := errors.New("mcp error")
		mockClient := &mockMCPClient{
			callToolFunc: func(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
				return nil, expectedErr
			},
		}

		toolProto := &v1.Tool{}
		toolProto.SetName("test-tool")
		toolProto.SetServiceId("service")
		mcpTool := tool.NewMCPTool(toolProto, mockClient, &configv1.MCPCallDefinition{})

		inputs := json.RawMessage(`{}`)
		req := &tool.ExecutionRequest{
			ToolName:   "service/-/test-tool",
			ToolInputs: inputs,
		}
		_, err := mcpTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), expectedErr.Error())
	})

	t.Run("with input transformation", func(t *testing.T) {
		mockClient := &mockMCPClient{
			callToolFunc: func(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
				args, ok := params.Arguments.(json.RawMessage)
				require.True(t, ok)
				assert.JSONEq(t, `{"transformed":true}`, string(args))
				return &mcp.CallToolResult{}, nil
			},
		}
		toolProto := &v1.Tool{}
		toolProto.SetName("test-tool")
		toolProto.SetServiceId("service")
		callDef := &configv1.MCPCallDefinition{}
		inputTransformer := &configv1.InputTransformer{}
		inputTransformer.SetTemplate(`{"transformed":true}`)
		callDef.SetInputTransformer(inputTransformer)
		mcpTool := tool.NewMCPTool(toolProto, mockClient, callDef)

		inputs := json.RawMessage(`{}`)
		req := &tool.ExecutionRequest{
			ToolName:   "service/-/test-tool",
			ToolInputs: inputs,
		}
		_, err := mcpTool.Execute(context.Background(), req)
		require.NoError(t, err)
	})
}
