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

package mcpserver_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/mcpxy/mcpx/pkg/auth"
	"github.com/mcpxy/mcpx/pkg/bus"
	"github.com/mcpxy/mcpx/pkg/mcpserver"
	"github.com/mcpxy/mcpx/pkg/pool"
	"github.com/mcpxy/mcpx/pkg/prompt"
	"github.com/mcpxy/mcpx/pkg/resource"
	"github.com/mcpxy/mcpx/pkg/serviceregistry"
	"github.com/mcpxy/mcpx/pkg/tool"
	"github.com/mcpxy/mcpx/pkg/upstream/factory"
	"github.com/mcpxy/mcpx/pkg/util"
	"github.com/mcpxy/mcpx/pkg/worker"
	v1 "github.com/mcpxy/mcpx/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
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

func TestToolValidationMiddleware(t *testing.T) {
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager)
	toolManager := tool.NewToolManager()
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	authManager := auth.NewAuthManager()
	serviceRegistry := serviceregistry.New(factory, toolManager, promptManager, resourceManager, authManager)
	busProvider := bus.NewBusProvider()
	ctx := context.Background()

	// Start the worker to handle tool execution
	upstreamWorker := worker.NewUpstreamWorker(busProvider, toolManager)
	upstreamWorker.Start(ctx)

	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider)
	require.NoError(t, err)

	tm := server.ToolManager().(*tool.ToolManager)

	// Add a test tool
	serviceKey := "test-service"
	toolName := "test-tool"
	compositeName, err := util.GenerateToolID(serviceKey, toolName)
	require.NoError(t, err)

	testTool := &mockTool{
		tool: v1.Tool_builder{
			Name:      proto.String(toolName),
			ServiceId: proto.String(serviceKey),
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

	// Test tools/call for a non-existent tool
	_, err = clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "non-existent-tool",
	})
	assert.ErrorContains(t, err, tool.ErrToolNotFound.Error())

	// Test tools/call for an existing tool
	_, err = clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: compositeName,
	})
	assert.NoError(t, err)

	// Test tools/list
	listResult, err := clientSession.ListTools(ctx, &mcp.ListToolsParams{})
	assert.NoError(t, err)
	assert.Len(t, listResult.Tools, 1)
	assert.Equal(t, compositeName, listResult.Tools[0].Name)

	// Remove the tool and test tools/list again
	tm.ClearToolsForService(serviceKey)
	listResult, err = clientSession.ListTools(ctx, &mcp.ListToolsParams{})
	assert.NoError(t, err)
	assert.Len(t, listResult.Tools, 0)
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

func TestServer_CallTool(t *testing.T) {
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager)
	toolManager := tool.NewToolManager()
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	authManager := auth.NewAuthManager()
	serviceRegistry := serviceregistry.New(factory, toolManager, promptManager, resourceManager, authManager)
	busProvider := bus.NewBusProvider()
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
		}.Build(),
	}
	tm.AddTool(successTool)

	errorTool := &mockErrorTool{
		tool: v1.Tool_builder{
			Name:      proto.String("error-tool"),
			ServiceId: proto.String("test-service"),
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
		toolID, _ := util.GenerateToolID("test-service", "success-tool")
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
		toolID, _ := util.GenerateToolID("test-service", "error-tool")
		_, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
			Name: toolID,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "execution error")
	})

	t.Run("tool call with context timeout", func(t *testing.T) {
		timeoutCtx, cancelTimeout := context.WithTimeout(ctx, 10*time.Millisecond)
		defer cancelTimeout()

		toolID, _ := util.GenerateToolID("test-service", "success-tool")
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
	toolManager := tool.NewToolManager()
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	authManager := auth.NewAuthManager()
	serviceRegistry := serviceregistry.New(factory, toolManager, promptManager, resourceManager, authManager)
	busProvider := bus.NewBusProvider()
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
	toolManager := tool.NewToolManager()
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	authManager := auth.NewAuthManager()
	serviceRegistry := serviceregistry.New(factory, toolManager, promptManager, resourceManager, authManager)
	busProvider := bus.NewBusProvider()
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
	toolManager := tool.NewToolManager()
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	authManager := auth.NewAuthManager()
	serviceRegistry := serviceregistry.New(factory, toolManager, promptManager, resourceManager, authManager)
	busProvider := bus.NewBusProvider()
	ctx := context.Background()

	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider)
	require.NoError(t, err)

	assert.NotNil(t, server.AuthManager())
	assert.NotNil(t, server.ToolManager())
	assert.NotNil(t, server.PromptManager())
	assert.NotNil(t, server.ResourceManager())
	assert.NotNil(t, server.ServiceRegistry())
}

func TestToolValidationMiddleware_Filter(t *testing.T) {
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager)
	toolManager := tool.NewToolManager()
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	authManager := auth.NewAuthManager()
	serviceRegistry := serviceregistry.New(factory, toolManager, promptManager, resourceManager, authManager)
	busProvider := bus.NewBusProvider()
	ctx := context.Background()

	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider)
	require.NoError(t, err)

	// Add a tool directly to the MCP server, but not to the tool manager
	mcpSDKTool := &mcp.Tool{
		Name:        "unmanaged-tool",
		InputSchema: &jsonschema.Schema{Type: "object"},
	}
	server.Server().AddTool(mcpSDKTool, func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return nil, nil
	})

	// Create client-server connection
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	// List tools, the unmanaged tool should be filtered out
	listResult, err := clientSession.ListTools(ctx, &mcp.ListToolsParams{})
	assert.NoError(t, err)
	assert.Len(t, listResult.Tools, 0)

	// Now add a managed tool
	managedTool := &mockTool{
		tool: v1.Tool_builder{
			Name:      proto.String("managed-tool"),
			ServiceId: proto.String("managed-service"),
		}.Build(),
	}
	toolManager.AddTool(managedTool)

	// List tools again, only the managed tool should be present
	listResult, err = clientSession.ListTools(ctx, &mcp.ListToolsParams{})
	assert.NoError(t, err)
	assert.Len(t, listResult.Tools, 1)
	managedToolID, _ := util.GenerateToolID("managed-service", "managed-tool")
	assert.Equal(t, managedToolID, listResult.Tools[0].Name)
}
