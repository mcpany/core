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

// featureSamplingTool is a mock tool that attempts to use the Sampling feature.
type featureSamplingTool struct {
	tool *v1.Tool
}

func (m *featureSamplingTool) Tool() *v1.Tool {
	return m.tool
}

func (m *featureSamplingTool) Execute(ctx context.Context, _ *tool.ExecutionRequest) (any, error) {
	// 1. Get Session/Sampler from context
	sampler, ok := tool.GetSession(ctx)
	if !ok {
		return "no sampler found", nil
	}

	// 2. Call CreateMessage
	res, err := sampler.CreateMessage(ctx, &mcp.CreateMessageParams{
		Messages: []*mcp.SamplingMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{
					Text: "hello from tool",
				},
			},
		},
		MaxTokens: 50,
	})
	if err != nil {
		return nil, err
	}

	// 3. Return the content from the Sampling response
	if tc, ok := res.Content.(*mcp.TextContent); ok {
		return tc.Text, nil
	}
	return "unknown content type", nil
}

func (m *featureSamplingTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

func (m *featureSamplingTool) MCPTool() *mcp.Tool {
	t, _ := tool.ConvertProtoToMCPTool(m.tool)
	return t
}

func TestFeatureSamplingSupport(t *testing.T) {
	// Setup Server Components
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

	// Initialize Server
	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider, false)
	require.NoError(t, err)

	tm := server.ToolManager().(*tool.Manager)

	// Register the Sampling Tool
	sTool := &featureSamplingTool{
		tool: v1.Tool_builder{
			Name:      proto.String("feature-sampling-tool"),
			ServiceId: proto.String("feature-service"),
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

	// Setup Client Mock
	clientImpl := &mcp.Implementation{Name: "feature-test-client"}
	samplingHandlerCalled := false

	// Define Client Options with CreateMessageHandler
	clientOpts := &mcp.ClientOptions{
		CreateMessageHandler: func(ctx context.Context, req *mcp.CreateMessageRequest) (*mcp.CreateMessageResult, error) {
			samplingHandlerCalled = true

			// Verify request content
			params := req.Params
			require.NotEmpty(t, params.Messages)
			if tc, ok := params.Messages[0].Content.(*mcp.TextContent); ok {
				assert.Equal(t, "hello from tool", tc.Text)
			}

			// Return mock response
			return &mcp.CreateMessageResult{
				Role: "assistant",
				Content: &mcp.TextContent{
					Text: "This is a sampled response from client",
				},
				Model: "mock-model-v1",
			}, nil
		},
	}

	client := mcp.NewClient(clientImpl, clientOpts)

	// Create InMemory Transport
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	// Connect Server and Client
	serverSession, err := server.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()

	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	// Execute the Tool via Client
	sanitizedToolName, _ := util.SanitizeToolName("feature-sampling-tool")
	toolID := "feature-service" + "." + sanitizedToolName

	// Verify tool is available
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

	// Call the Tool
	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: toolID,
	})
	require.NoError(t, err)

	// Verify Result
	require.Len(t, result.Content, 1)
	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)

	// Tool returns result as JSON string (usually) or direct string if simple
	// In this case, our tool returns a string.
	// But `CallTool` wraps it in `CallToolResult`.
	// Our `Execute` method returns `string`.
	// `mcpserver.CallTool` wraps string in `TextContent`.
	// The content of `TextContent` is the string.

	// Wait, `mcpserver.CallTool` implementation:
	// return &mcp.CallToolResult{ Content: ... Text: util.ToString(res) }
	// So it should be the raw string.

	// Check if it's JSON encoded (standard behavior for objects, but strings might be raw)
	// `util.ToString` usually returns raw string.
	// However, `ExecuteTool` might return `mcp.CallToolResult`.
	// In `ExecuteTool` (manager), it returns `any`.
	// In `Server.CallTool`:
	// if result is not `mcp.CallToolResult`, it wraps it.
	// The wrapper uses `util.ToString`.

	// Let's check `util.ToString` behavior?
	// Assuming it returns "This is a sampled response from client".

	// The previous test `sampling_test.go` used `json.Unmarshal`.
	// `var resString string; err = json.Unmarshal([]byte(textContent.Text), &resString)`
	// This implies `textContent.Text` was `"This is a sampled response"`.
	// Why quoted? `util.ToString` might use JSON marshaling for some types or `fmt.Sprint`.
	// If `Execute` returns `string`, `util.ToString` likely returns it as is?
	// Wait, `server.go`:
	// `jsonBytes, marshalErr = json.Marshal(result)` -> "string" -> `"string"`
	// So yes, it is JSON quoted string.

	var resString string
	err = json.Unmarshal([]byte(textContent.Text), &resString)
	require.NoError(t, err)

	assert.Equal(t, "This is a sampled response from client", resString)
	assert.True(t, samplingHandlerCalled, "CreateMessage handler was not called on client")
}
