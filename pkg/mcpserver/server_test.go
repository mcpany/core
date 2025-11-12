
package mcpserver_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

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
	"github.com/mcpany/core/pkg/worker"
	bus_pb "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type mockTool struct {
	tool *v1.Tool
}

func (m *mockTool) Tool() *v1.Tool {
	return m.tool
}

func (m *mockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	// Simulate work that takes a bit of time, allowing context cancellation to be tested.
	select {
	case <-time.After(50 * time.Millisecond):
		return "success", nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (m *mockTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

func TestToolListFiltering(t *testing.T) {
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager)
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewBusProvider(messageBus)
	require.NoError(t, err)
	toolManager := tool.NewToolManager(busProvider)
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	authManager := auth.NewAuthManager()
	serviceRegistry := serviceregistry.New(factory, toolManager, promptManager, resourceManager, authManager)
	ctx := context.Background()

	// Start the worker to handle tool execution
	upstreamWorker := worker.NewUpstreamWorker(busProvider, toolManager)
	upstreamWorker.Start(ctx)

	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider)
	require.NoError(t, err)

	tm := server.ToolManager().(*tool.ToolManager)

	// Add a test tool
	serviceID := "test-service"
	toolName := "test-tool"
	sanitizedToolName, err := util.SanitizeToolName(toolName)
	require.NoError(t, err)
	compositeName := serviceID + "." + sanitizedToolName

	testTool := &mockTool{
		tool: v1.Tool_builder{
			Name:      proto.String(toolName),
			ServiceId: proto.String(serviceID),
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
	tm.AddTool(testTool)

	// Create client-server connection for demonstration
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	// Connect server and client
	serverSession, err := server.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	// Test tools/list
	listResult, err := clientSession.ListTools(ctx, &mcp.ListToolsParams{})
	assert.NoError(t, err)
	assert.Len(t, listResult.Tools, 1)
	assert.Equal(t, compositeName, listResult.Tools[0].Name)

	// Remove the tool and test tools/list again
	tm.ClearToolsForService(serviceID)
	listResult, err = clientSession.ListTools(ctx, &mcp.ListToolsParams{})
	assert.NoError(t, err)
	assert.Len(t, listResult.Tools, 0)
}

func TestToolListFilteringServiceId(t *testing.T) {
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager)
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewBusProvider(messageBus)
	require.NoError(t, err)
	toolManager := tool.NewToolManager(busProvider)
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	authManager := auth.NewAuthManager()
	serviceRegistry := serviceregistry.New(factory, toolManager, promptManager, resourceManager, authManager)
	ctx := context.Background()

	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider)
	require.NoError(t, err)

	tm := server.ToolManager().(*tool.ToolManager)

	// Add a test tool
	serviceID := "test-service"
	toolName := "test-tool"
	sanitizedToolName, err := util.SanitizeToolName(toolName)
	require.NoError(t, err)
	compositeName := serviceID + "." + sanitizedToolName

	testTool := &mockTool{
		tool: v1.Tool_builder{
			Name:      proto.String(toolName),
			ServiceId: proto.String(serviceID),
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
	tm.AddTool(testTool)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	serverOpts := &mcp.ServerSessionOptions{
		State: &mcp.ServerSessionState{
			InitializeParams: &mcp.InitializeParams{
				Capabilities: &mcp.ClientCapabilities{},
			},
		},
	}

	serverSession, err := server.Server().Connect(ctx, serverTransport, serverOpts)
	require.NoError(t, err)
	defer serverSession.Close()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	listResult, err := clientSession.ListTools(ctx, &mcp.ListToolsParams{})
	assert.NoError(t, err)
	assert.Len(t, listResult.Tools, 1)
	assert.Equal(t, compositeName, listResult.Tools[0].Name)
}

type mockErrorTool struct {
	tool *v1.Tool
}

func (m *mockErrorTool) Tool() *v1.Tool {
	return m.tool
}

func (m *mockErrorTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return nil, errors.New("execution error")
}

func (m *mockErrorTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

func TestServer_CallTool(t *testing.T) {
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager)
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewBusProvider(messageBus)
	require.NoError(t, err)
	toolManager := tool.NewToolManager(busProvider)
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	authManager := auth.NewAuthManager()
	serviceRegistry := serviceregistry.New(factory, toolManager, promptManager, resourceManager, authManager)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the worker to handle tool execution
	upstreamWorker := worker.NewUpstreamWorker(busProvider, toolManager)
	upstreamWorker.Start(ctx)

	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider)
	require.NoError(t, err)

	tm := server.ToolManager().(*tool.ToolManager)

	// Add test tools
	successTool := &mockTool{
		tool: v1.Tool_builder{
			Name:      proto.String("success-tool"),
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
	tm.AddTool(successTool)

	errorTool := &mockErrorTool{
		tool: v1.Tool_builder{
			Name:      proto.String("error-tool"),
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
	tm.AddTool(errorTool)

	// Create client-server connection
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	// Connect server and client
	serverSession, err := server.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	t.Run("successful tool call", func(t *testing.T) {
		sanitizedToolName, _ := util.SanitizeToolName("success-tool")
		toolID := "test-service" + "." + sanitizedToolName
		result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
			Name: toolID,
		})
		require.NoError(t, err)
		require.Len(t, result.Content, 1)
		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)
		var res string
		err = json.Unmarshal([]byte(textContent.Text), &res)
		require.NoError(t, err)
		assert.Equal(t, "success", res)
	})

	t.Run("tool call with error", func(t *testing.T) {
		sanitizedToolName, _ := util.SanitizeToolName("error-tool")
		toolID := "test-service" + "." + sanitizedToolName
		_, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
			Name: toolID,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "execution error")
	})

	t.Run("tool call with context timeout", func(t *testing.T) {
		timeoutCtx, cancelTimeout := context.WithTimeout(ctx, 10*time.Millisecond)
		defer cancelTimeout()

		sanitizedToolName, _ := util.SanitizeToolName("success-tool")
		toolID := "test-service" + "." + sanitizedToolName
		_, err := clientSession.CallTool(timeoutCtx, &mcp.CallToolParams{
			Name: toolID,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})
}

type testPrompt struct {
	NameValue    string
	ServiceValue string
}

func (p *testPrompt) Prompt() *mcp.Prompt {
	return &mcp.Prompt{Name: p.NameValue}
}

func (p *testPrompt) Service() string {
	return p.ServiceValue
}

func (p *testPrompt) Get(ctx context.Context, args json.RawMessage) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{}, nil
}

func TestServer_Prompts(t *testing.T) {
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager)
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewBusProvider(messageBus)
	require.NoError(t, err)
	toolManager := tool.NewToolManager(busProvider)
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	authManager := auth.NewAuthManager()
	serviceRegistry := serviceregistry.New(factory, toolManager, promptManager, resourceManager, authManager)
	ctx := context.Background()

	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider)
	require.NoError(t, err)

	// Add a test prompt
	promptManager.AddPrompt(&testPrompt{
		NameValue: "test-prompt",
	})

	// Create client-server connection
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	// Connect server and client
	serverSession, err := server.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	t.Run("list prompts", func(t *testing.T) {
		result, err := clientSession.ListPrompts(ctx, &mcp.ListPromptsParams{})
		require.NoError(t, err)
		require.Len(t, result.Prompts, 1)
		assert.Equal(t, "test-prompt", result.Prompts[0].Name)
	})

	t.Run("get prompt", func(t *testing.T) {
		_, err := clientSession.GetPrompt(ctx, &mcp.GetPromptParams{
			Name: "test-prompt",
		})
		require.NoError(t, err)
	})

	t.Run("get non-existent prompt", func(t *testing.T) {
		_, err := clientSession.GetPrompt(ctx, &mcp.GetPromptParams{
			Name: "non-existent-prompt",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), prompt.ErrPromptNotFound.Error())
	})
}

type testResource struct {
	URIValue     string
	ServiceValue string
}

func (r *testResource) Resource() *mcp.Resource {
	return &mcp.Resource{URI: r.URIValue}
}

func (r *testResource) Service() string {
	return r.ServiceValue
}

func (r *testResource) Read(ctx context.Context) (*mcp.ReadResourceResult, error) {
	return &mcp.ReadResourceResult{}, nil
}

func (r *testResource) Subscribe(ctx context.Context) error {
	return nil
}

func TestServer_Resources(t *testing.T) {
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager)
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewBusProvider(messageBus)
	require.NoError(t, err)
	toolManager := tool.NewToolManager(busProvider)
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	authManager := auth.NewAuthManager()
	serviceRegistry := serviceregistry.New(factory, toolManager, promptManager, resourceManager, authManager)
	ctx := context.Background()

	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider)
	require.NoError(t, err)

	// Add a test resource
	resourceManager.AddResource(&testResource{
		URIValue: "test-resource",
	})

	// Create client-server connection
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	// Connect server and client
	serverSession, err := server.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	t.Run("list resources", func(t *testing.T) {
		result, err := clientSession.ListResources(ctx, &mcp.ListResourcesParams{})
		require.NoError(t, err)
		require.Len(t, result.Resources, 1)
		assert.Equal(t, "test-resource", result.Resources[0].URI)
	})

	t.Run("read resource", func(t *testing.T) {
		_, err := clientSession.ReadResource(ctx, &mcp.ReadResourceParams{
			URI: "test-resource",
		})
		require.NoError(t, err)
	})

	t.Run("read non-existent resource", func(t *testing.T) {
		_, err := clientSession.ReadResource(ctx, &mcp.ReadResourceParams{
			URI: "non-existent-resource",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), resource.ErrResourceNotFound.Error())
	})
}

func TestServer_Getters(t *testing.T) {
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager)
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewBusProvider(messageBus)
	require.NoError(t, err)
	toolManager := tool.NewToolManager(busProvider)
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	authManager := auth.NewAuthManager()
	serviceRegistry := serviceregistry.New(factory, toolManager, promptManager, resourceManager, authManager)
	ctx := context.Background()

	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider)
	require.NoError(t, err)

	assert.NotNil(t, server.AuthManager())
	assert.NotNil(t, server.ToolManager())
	assert.NotNil(t, server.PromptManager())
	assert.NotNil(t, server.ResourceManager())
	assert.NotNil(t, server.ServiceRegistry())
}

type mockToolManager struct {
	tool.ToolManager
	addServiceInfoCalled       bool
	getToolCalled              bool
	listToolsCalled            bool
	executeToolCalled          bool
	setMCPServerCalled         bool
	addToolCalled              bool
	getServiceInfoCalled       bool
	clearToolsForServiceCalled bool
}

func (m *mockToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	m.addServiceInfoCalled = true
}

func (m *mockToolManager) GetTool(toolName string) (tool.Tool, bool) {
	m.getToolCalled = true
	return &mockTool{}, true
}

func (m *mockToolManager) ListTools() []tool.Tool {
	m.listToolsCalled = true
	return []tool.Tool{}
}

func (m *mockToolManager) ExecuteTool(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	m.executeToolCalled = true
	return nil, nil
}

func (m *mockToolManager) SetMCPServer(mcpServer tool.MCPServerProvider) {
	m.setMCPServerCalled = true
}

func (m *mockToolManager) AddTool(t tool.Tool) error {
	m.addToolCalled = true
	return nil
}

func (m *mockToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	m.getServiceInfoCalled = true
	return &tool.ServiceInfo{}, true
}

func (m *mockToolManager) ClearToolsForService(serviceKey string) {
	m.clearToolsForServiceCalled = true
}

func TestServer_ToolManagerDelegation(t *testing.T) {
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager)
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewBusProvider(messageBus)
	require.NoError(t, err)
	mockToolManager := &mockToolManager{}
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	authManager := auth.NewAuthManager()
	serviceRegistry := serviceregistry.New(factory, mockToolManager, promptManager, resourceManager, authManager)
	ctx := context.Background()

	server, err := mcpserver.NewServer(ctx, mockToolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider)
	require.NoError(t, err)

	server.AddServiceInfo("test-service", &tool.ServiceInfo{})
	assert.True(t, mockToolManager.addServiceInfoCalled)

	_, _ = server.GetTool("test-tool")
	assert.True(t, mockToolManager.getToolCalled)

	_ = server.ListTools()
	assert.True(t, mockToolManager.listToolsCalled)

	_, _ = server.ExecuteTool(ctx, &tool.ExecutionRequest{})
	assert.True(t, mockToolManager.executeToolCalled)

	server.SetMCPServer(nil)
	assert.True(t, mockToolManager.setMCPServerCalled)

	_ = server.AddTool(&mockTool{})
	assert.True(t, mockToolManager.addToolCalled)

	_, _ = server.GetServiceInfo("test-service")
	assert.True(t, mockToolManager.getServiceInfoCalled)

	server.ClearToolsForService("test-service")
	assert.True(t, mockToolManager.clearToolsForServiceCalled)
}
