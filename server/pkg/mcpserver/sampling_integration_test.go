// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/mcpserver"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/sampling"
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

// mockSamplingTool is a tool that requests sampling
type mockSamplingTool struct {
	tool *v1.Tool
}

func (m *mockSamplingTool) Tool() *v1.Tool {
	return m.tool
}

func (m *mockSamplingTool) Execute(ctx context.Context, _ *tool.ExecutionRequest) (any, error) {
	// Get Sampler from context
	sampler, ok := sampling.FromContext(ctx)
	if !ok {
		return "error: no sampler in context", nil
	}

	// Create Message
	result, err := sampler.CreateMessage(ctx, &mcp.CreateMessageParams{
		Messages: []*mcp.SamplingMessage{
			{
				Role: mcp.Role("user"),
				Content: &mcp.TextContent{
					Text: "hello",
				},
			},
		},
		MaxTokens: 100,
	})

	if err != nil {
		return "error: " + err.Error(), nil
	}

	// Return the text content from the result
	if tc, ok := result.Content.(*mcp.TextContent); ok {
		return tc.Text, nil
	}
	return "error: unexpected content type", nil
}

func (m *mockSamplingTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

func (m *mockSamplingTool) MCPTool() *mcp.Tool {
	t, _ := tool.ConvertProtoToMCPTool(m.tool)
	return t
}

func TestSamplingIntegration(t *testing.T) {
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider, false)
	require.NoError(t, err)

	tm := server.ToolManager().(*tool.Manager)

	// Register Sampling Tool
	toolName := "sampling-tool"
	testTool := &mockSamplingTool{
		tool: v1.Tool_builder{
			Name:      proto.String(toolName),
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
	_ = tm.AddTool(testTool)

	// Create Client with Sampling Handler
	clientHandler := func(ctx context.Context, req *mcp.ClientRequest[*mcp.CreateMessageParams]) (*mcp.CreateMessageResult, error) {
		return &mcp.CreateMessageResult{
			Role: mcp.Role("assistant"),
			Content: &mcp.TextContent{
				Text: "I am a mocked LLM",
			},
			Model: "mock-model",
		}, nil
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, &mcp.ClientOptions{
		CreateMessageHandler: clientHandler,
	})

	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	// Connect
	serverSession, err := server.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()

	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	// Call the tool from Client
	sanitizedToolName, _ := util.SanitizeToolName(toolName)
	fullToolName := "test-service." + sanitizedToolName

	callResult, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: fullToolName,
	})
	require.NoError(t, err)

	// Verify result
	require.Len(t, callResult.Content, 1)
	textContent, ok := callResult.Content[0].(*mcp.TextContent)
	require.True(t, ok)

	var resultStr string
	// Tool returns string, which is JSON encoded by CallTool into TextContent.
	err = json.Unmarshal([]byte(textContent.Text), &resultStr)
	require.NoError(t, err)

	assert.Equal(t, "I am a mocked LLM", resultStr)
}
