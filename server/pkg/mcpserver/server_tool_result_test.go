// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver_test

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/mcpserver"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/serviceregistry"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream/factory"
	"github.com/mcpany/core/pkg/util"
	bus_pb "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type mockMapResultTool struct {
	name   string
	result map[string]any
}

func (m *mockMapResultTool) Tool() *v1.Tool {
	return &v1.Tool{
		Name:      proto.String(m.name),
		ServiceId: proto.String("test-service"),
		Annotations: &v1.ToolAnnotations{
			InputSchema: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"type":       structpb.NewStringValue("object"),
					"properties": structpb.NewStructValue(&structpb.Struct{}),
				},
			},
		},
	}
}

func (m *mockMapResultTool) Execute(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return m.result, nil
}

func (m *mockMapResultTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

func (m *mockMapResultTool) MCPTool() *mcp.Tool {
	t, _ := tool.ConvertProtoToMCPTool(m.Tool())
	return t
}

func TestServer_CallTool_ResultHandling(t *testing.T) {
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager)
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(t, err)
	toolManager := tool.NewManager(busProvider)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(factory, toolManager, promptManager, resourceManager, authManager)
	ctx := context.Background()

	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider, false)
	require.NoError(t, err)

	tm := server.ToolManager().(*tool.Manager)

	// Tool 1: "Smart" map result (looks like CallToolResult)
	smartTool := &mockMapResultTool{
		name: "smart-tool",
		result: map[string]any{
			"content": []any{
				map[string]any{
					"type": "text",
					"text": "hello",
				},
			},
			"isError": false,
		},
	}
	_ = tm.AddTool(smartTool)

	// Tool 2: "Ambiguous" map result (has content field but not CallToolResult structure)
	ambiguousTool := &mockMapResultTool{
		name: "ambiguous-tool",
		result: map[string]any{
			"content": "some raw text content", // String, not array of Content objects
		},
	}
	_ = tm.AddTool(ambiguousTool)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	t.Run("Smart Result Handling", func(t *testing.T) {
		sanitizedToolName, _ := util.SanitizeToolName("smart-tool")
		toolID := "test-service" + "." + sanitizedToolName
		result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
			Name: toolID,
		})
		require.NoError(t, err)

		require.Len(t, result.Content, 1)
		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)

		// Should NOT be JSON stringified
		if textContent.Text != "hello" {
			t.Fatalf("Expected smart logic (TextContent with 'hello'), but got: %s", textContent.Text)
		}
	})

	t.Run("Ambiguous Result Fallback", func(t *testing.T) {
		sanitizedToolName, _ := util.SanitizeToolName("ambiguous-tool")
		toolID := "test-service" + "." + sanitizedToolName
		result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
			Name: toolID,
		})
		require.NoError(t, err)

		require.Len(t, result.Content, 1)
		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)

		// It should be the content text directly (improved fallback behavior)
		assert.Equal(t, "some raw text content", textContent.Text)
	})
}
