// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/mcpany/core/server/pkg/client"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHTTPTool_InitError(t *testing.T) {
	// Setup with invalid template to trigger initError
	callDef := &configv1.HttpCallDefinition{
		InputTransformer: &configv1.InputTransformer{
			Template: proto.String("{{bad-template"),
		},
	}
	// Need valid tool proto
	toolProto := &v1.Tool{
		Name:                proto.String("test"),
		UnderlyingMethodFqn: proto.String("GET http://example.com"),
	}

	pm := pool.NewManager()
    // Register dummy pool to pass the pool check
    p, err := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
        return &client.HTTPClientWrapper{}, nil
    }, 1, 1, 0, true)
    require.NoError(t, err)
    pm.Register("service", p)

	httpTool := tool.NewHTTPTool(toolProto, pm, "service", nil, callDef, nil, nil, "call-id")

	req := &tool.ExecutionRequest{
		ToolName: "test",
		ToolInputs: []byte("{}"),
	}
	_, err = httpTool.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse input template")
}

func TestHTTPTool_DebugLogging(t *testing.T) {
	// Reset logger to capture debug logs (or just ensure debug is on)
	logging.ForTestsOnlyResetLogger()
	logging.Init(slog.LevelDebug, io.Discard)

	callDef := &configv1.HttpCallDefinition{}
	toolProto := &v1.Tool{
		Name:                proto.String("test"),
		UnderlyingMethodFqn: proto.String("GET http://example.com"),
	}

	pm := pool.NewManager()

	httpTool := tool.NewHTTPTool(toolProto, pm, "service", nil, callDef, nil, nil, "call-id")

	req := &tool.ExecutionRequest{
		ToolName: "test",
		ToolInputs: []byte("{}"),
	}
	// This will fail because no pool found, but it should hit debug logging "executing tool" first.
	_, err := httpTool.Execute(context.Background(), req)
	assert.Error(t, err)
}

func TestLocalCommandTool_DebugLogging(t *testing.T) {
	logging.ForTestsOnlyResetLogger()
	logging.Init(slog.LevelDebug, io.Discard)

	service := &configv1.CommandLineUpstreamService{
		Command: proto.String("echo"),
        Local: proto.Bool(true),
	}
	callDef := &configv1.CommandLineCallDefinition{}
	toolProto := &v1.Tool{Name: proto.String("echo")}

	cmdTool := tool.NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	req := &tool.ExecutionRequest{
		ToolName: "echo",
		ToolInputs: []byte("{}"),
	}
	// Should work (echo is local)
	_, err := cmdTool.Execute(context.Background(), req)
	assert.NoError(t, err)
}

func TestOpenAPITool_InitError(t *testing.T) {
	callDef := &configv1.OpenAPICallDefinition{
		InputTransformer: &configv1.InputTransformer{
			Template: proto.String("{{bad-template"),
		},
	}
	toolProto := &v1.Tool{Name: proto.String("test")}

	openapiTool := tool.NewOpenAPITool(toolProto, nil, nil, "GET", "http://example.com", nil, callDef)

	req := &tool.ExecutionRequest{
		ToolName: "test",
		ToolInputs: []byte("{}"),
	}
	_, err := openapiTool.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse input template")
}

func TestMCPTool_InitError(t *testing.T) {
	callDef := &configv1.MCPCallDefinition{
		InputTransformer: &configv1.InputTransformer{
			Template: proto.String("{{bad-template"),
		},
	}
	toolProto := &v1.Tool{Name: proto.String("test")}

	mcpTool := tool.NewMCPTool(toolProto, nil, callDef)

	req := &tool.ExecutionRequest{
		ToolName: "test",
		ToolInputs: []byte("{}"),
	}
	_, err := mcpTool.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse input template")
}

func TestMCPTool_DebugLogging(t *testing.T) {
	logging.ForTestsOnlyResetLogger()
	logging.Init(slog.LevelDebug, io.Discard)

	callDef := &configv1.MCPCallDefinition{}
	toolProto := &v1.Tool{Name: proto.String("test")}

    // Use the existing mock from mcp_tool_test.go

    mock := &mockMCPClient{
        callToolFunc: func(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
            return &mcp.CallToolResult{
                Content: []mcp.Content{
                    &mcp.TextContent{Text: "result"},
                },
            }, nil
        },
    }

	mcpTool := tool.NewMCPTool(toolProto, mock, callDef)

	req := &tool.ExecutionRequest{
		ToolName: "test",
		ToolInputs: []byte("{}"),
	}
	_, err := mcpTool.Execute(context.Background(), req)
	assert.NoError(t, err)
}
