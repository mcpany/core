// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/mcpserver"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/mcpany/core/server/pkg/util"
	bus_pb "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type samplingTool struct {
	tool *v1.Tool
}

func (m *samplingTool) Tool() *v1.Tool {
	return m.tool
}

func (m *samplingTool) Execute(ctx context.Context, _ *tool.ExecutionRequest) (any, error) {
	sampler, ok := tool.GetSampler(ctx)
	if !ok {
		return "no sampler found", nil
	}

	res, err := sampler.CreateMessage(ctx, &mcp.CreateMessageParams{
		Messages: []*mcp.SamplingMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{
					Text: "hello",
				},
			},
		},
		MaxTokens: 10,
	})
	if err != nil {
		return nil, err
	}

	// Check content type
	if tc, ok := res.Content.(*mcp.TextContent); ok {
		return tc.Text, nil
	}
	return "unknown content type", nil
}

func (m *samplingTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

func (m *samplingTool) MCPTool() *mcp.Tool {
	t, _ := tool.ConvertProtoToMCPTool(m.tool)
	return t
}

func TestSamplingSupport(t *testing.T) {
	// Setup Server
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager, nil)
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

	// Add sampling tool
	sTool := &samplingTool{
		tool: v1.Tool_builder{
			Name:      proto.String("sampling-tool"),
			ServiceId: proto.String("test-service"),
			Annotations: v1.ToolAnnotations_builder{
				InputSchema: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"type":       structpb.NewStringValue("object"),
						"properties": structpb.NewStructValue(&structpb.Struct{}),
					},
				},
			}.Build(),
		}.Build(),
	}
	err = tm.AddTool(sTool)
	require.NoError(t, err)

	// Setup Client with Sampling Capability
	clientImpl := &mcp.Implementation{Name: "test-client"}

	samplingHandlerCalled := false

	client := mcp.NewClient(clientImpl, &mcp.ClientOptions{
		CreateMessageHandler: func(ctx context.Context, req *mcp.CreateMessageRequest) (*mcp.CreateMessageResult, error) {
			samplingHandlerCalled = true
			params := req.Params
			if len(params.Messages) > 0 {
				if tc, ok := params.Messages[0].Content.(*mcp.TextContent); ok {
					assert.Equal(t, "hello", tc.Text)
				}
			}

			return &mcp.CreateMessageResult{
				Role: "assistant",
				Content: &mcp.TextContent{
					Text: "This is a sampled response",
				},
				Model: "mock-model",
			}, nil
		},
	})

	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	// Connect server and client
	serverSession, err := server.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()

	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	// Call the tool
	sanitizedToolName, _ := util.SanitizeToolName("sampling-tool")
	toolID := "test-service" + "." + sanitizedToolName

	// Ensure tool is listed
	listRes, err := clientSession.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	found := false
	for _, t := range listRes.Tools {
		if t.Name == toolID {
			found = true
			break
		}
	}
	require.True(t, found, "sampling tool not found in list")

	// Call it
	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: toolID,
	})
	require.NoError(t, err)

	require.Len(t, result.Content, 1)
	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)

	var resString string
	err = json.Unmarshal([]byte(textContent.Text), &resString)
	require.NoError(t, err)

	assert.Equal(t, "This is a sampled response", resString)
	assert.True(t, samplingHandlerCalled, "CreateMessage handler was not called")
}
