package tool_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/server/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
    "github.com/samber/lo"
)

func TestMCPTool_Execute_Coverage(t *testing.T) {
	t.Parallel()

    // Test result with no content
    t.Run("no_content", func(t *testing.T) {
        mockClient := &mockMCPClient{
            callToolFunc: func(_ context.Context, _ *mcp.CallToolParams) (*mcp.CallToolResult, error) {
                return &mcp.CallToolResult{Content: []mcp.Content{}}, nil
            },
        }

        toolProto := &v1.Tool{}
        toolProto.SetName("test-tool")
        mcpTool := tool.NewMCPTool(toolProto, mockClient, &configv1.MCPCallDefinition{})

        inputs := json.RawMessage(`{}`)
        req := &tool.ExecutionRequest{ToolInputs: inputs}
        result, err := mcpTool.Execute(context.Background(), req)
        require.NoError(t, err)
        assert.Nil(t, result)
    })

    // Test non-text content fallback
    t.Run("non_text_content", func(t *testing.T) {
        mockClient := &mockMCPClient{
            callToolFunc: func(_ context.Context, _ *mcp.CallToolParams) (*mcp.CallToolResult, error) {
                return &mcp.CallToolResult{
                    Content: []mcp.Content{
                        &mcp.ImageContent{Data: []byte("fake-image"), MIMEType: "image/png"},
                    },
                }, nil
            },
        }

        toolProto := &v1.Tool{}
        toolProto.SetName("test-tool")
        mcpTool := tool.NewMCPTool(toolProto, mockClient, &configv1.MCPCallDefinition{})

        inputs := json.RawMessage(`{}`)
        req := &tool.ExecutionRequest{ToolInputs: inputs}
        result, err := mcpTool.Execute(context.Background(), req)
        require.NoError(t, err)

        // Content is marshaled to JSON string and returned if it cannot be unmarshaled into a map.
        assert.IsType(t, "", result)
    })

    // Test output transformation: Raw Bytes
    t.Run("output_transformer_raw_bytes", func(t *testing.T) {
        mockClient := &mockMCPClient{
            callToolFunc: func(_ context.Context, _ *mcp.CallToolParams) (*mcp.CallToolResult, error) {
                return &mcp.CallToolResult{
                    Content: []mcp.Content{
                        &mcp.TextContent{Text: "some raw text"},
                    },
                }, nil
            },
        }

        toolProto := &v1.Tool{}
        toolProto.SetName("test-tool")
        format := configv1.OutputTransformer_RAW_BYTES
        callDef := configv1.MCPCallDefinition_builder{
            OutputTransformer: configv1.OutputTransformer_builder{
                Format: &format,
            }.Build(),
        }.Build()

        mcpTool := tool.NewMCPTool(toolProto, mockClient, callDef)

        inputs := json.RawMessage(`{}`)
        req := &tool.ExecutionRequest{ToolInputs: inputs}
        result, err := mcpTool.Execute(context.Background(), req)
        require.NoError(t, err)

        resultMap, ok := result.(map[string]any)
        require.True(t, ok)
        assert.Equal(t, []byte("some raw text"), resultMap["raw"])
    })

    // Test output transformation: JQ
    t.Run("output_transformer_jq", func(t *testing.T) {
        mockClient := &mockMCPClient{
            callToolFunc: func(_ context.Context, _ *mcp.CallToolParams) (*mcp.CallToolResult, error) {
                return &mcp.CallToolResult{
                    Content: []mcp.Content{
                        &mcp.TextContent{Text: `{"key": "value"}`},
                    },
                }, nil
            },
        }

        toolProto := &v1.Tool{}
        toolProto.SetName("test-tool")
        format := configv1.OutputTransformer_JQ
        callDef := configv1.MCPCallDefinition_builder{
            OutputTransformer: configv1.OutputTransformer_builder{
                Format: &format,
                JqQuery: lo.ToPtr(".key"),
            }.Build(),
        }.Build()

        mcpTool := tool.NewMCPTool(toolProto, mockClient, callDef)

        inputs := json.RawMessage(`{}`)
        req := &tool.ExecutionRequest{ToolInputs: inputs}
        result, err := mcpTool.Execute(context.Background(), req)
        require.NoError(t, err)
        assert.Equal(t, "value", result)
    })

    // Test output transformation: Template
    t.Run("output_transformer_template", func(t *testing.T) {
        mockClient := &mockMCPClient{
            callToolFunc: func(_ context.Context, _ *mcp.CallToolParams) (*mcp.CallToolResult, error) {
                return &mcp.CallToolResult{
                    Content: []mcp.Content{
                        &mcp.TextContent{Text: `{"key": "value"}`},
                    },
                }, nil
            },
        }

        toolProto := &v1.Tool{}
        toolProto.SetName("test-tool")
        format := configv1.OutputTransformer_JSON
        callDef := configv1.MCPCallDefinition_builder{
            OutputTransformer: configv1.OutputTransformer_builder{
                Format: &format,
                ExtractionRules: map[string]string{"key": "{.key}"},
                Template: lo.ToPtr("Value is {{key}}"),
            }.Build(),
        }.Build()

        mcpTool := tool.NewMCPTool(toolProto, mockClient, callDef)

        inputs := json.RawMessage(`{}`)
        req := &tool.ExecutionRequest{ToolInputs: inputs}
        result, err := mcpTool.Execute(context.Background(), req)
        require.NoError(t, err)

        resultMap, ok := result.(map[string]any)
        require.True(t, ok)
        assert.Equal(t, "Value is value", resultMap["result"])
    })

    // Test input decode error
    t.Run("input_decode_error", func(t *testing.T) {
        toolProto := &v1.Tool{}
        toolProto.SetName("test-tool")
        mcpTool := tool.NewMCPTool(toolProto, &mockMCPClient{}, &configv1.MCPCallDefinition{})

        inputs := json.RawMessage(`{"invalid":`) // Invalid JSON
        req := &tool.ExecutionRequest{ToolInputs: inputs}
        _, err := mcpTool.Execute(context.Background(), req)
        require.Error(t, err)
        assert.Contains(t, err.Error(), "failed to unmarshal tool inputs")
    })
}
