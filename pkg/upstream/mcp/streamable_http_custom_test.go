/*
 * Copyright 2025 Author(s) of MCP Any
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

package mcp

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/mcpany/core/pkg/client"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// MockClientSession is a mock implementation of the ClientSession interface
type MockClientSession struct {
	mock.Mock
}

func (m *MockClientSession) ListTools(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mcp.ListToolsResult), args.Error(1)
}

func (m *MockClientSession) ListPrompts(ctx context.Context, params *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mcp.ListPromptsResult), args.Error(1)
}

func (m *MockClientSession) ListResources(ctx context.Context, params *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mcp.ListResourcesResult), args.Error(1)
}

func (m *MockClientSession) GetPrompt(ctx context.Context, params *mcp.GetPromptParams) (*mcp.GetPromptResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mcp.GetPromptResult), args.Error(1)
}

func (m *MockClientSession) ReadResource(ctx context.Context, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mcp.ReadResourceResult), args.Error(1)
}

func (m *MockClientSession) CallTool(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mcp.CallToolResult), args.Error(1)
}

func (m *MockClientSession) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestMcpConnection_CallTool(t *testing.T) {
	conn := &mcpConnection{
		stdioConfig: &configv1.McpStdioConnection{},
	}
	ctx := context.Background()
	params := &mcp.CallToolParams{Name: "test-tool"}

	t.Run("successful call", func(t *testing.T) {
		expectedResult := &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "success"}}}
		originalConnect := connectForTesting
		SetConnectForTesting(func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
			mockSession := new(MockClientSession)
			mockSession.On("CallTool", ctx, params).Return(expectedResult, nil)
			mockSession.On("Close").Return(nil)
			return mockSession, nil
		})
		defer func() { connectForTesting = originalConnect }()

		result, err := conn.CallTool(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("connection error", func(t *testing.T) {
		connectErr := errors.New("connection failed")
		originalConnect := connectForTesting
		SetConnectForTesting(func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
			return nil, connectErr
		})
		defer func() { connectForTesting = originalConnect }()

		_, err := conn.CallTool(ctx, params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), connectErr.Error())
	})

	t.Run("tool call error", func(t *testing.T) {
		toolErr := errors.New("tool call failed")
		originalConnect := connectForTesting
		SetConnectForTesting(func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
			mockSession := new(MockClientSession)
			mockSession.On("CallTool", ctx, params).Return(nil, toolErr)
			mockSession.On("Close").Return(nil)
			return mockSession, nil
		})
		defer func() { connectForTesting = originalConnect }()

		_, err := conn.CallTool(ctx, params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), toolErr.Error())
	})
}

func TestSetTestingHooks(t *testing.T) {
	t.Run("SetNewClientImplForTesting", func(t *testing.T) {
		var called bool
		SetNewClientImplForTesting(func(client *mcp.Client, stdioConfig *configv1.McpStdioConnection, httpAddress string, httpClient *http.Client) client.MCPClient {
			called = true
			return nil
		})
		assert.NotNil(t, newClientImplForTesting)
		newClientImplForTesting(nil, nil, "", nil)
		assert.True(t, called)
		newClientImplForTesting = nil // Reset for other tests
	})

	t.Run("SetNewClientForTesting", func(t *testing.T) {
		var called bool
		SetNewClientForTesting(func(impl *mcp.Implementation) *mcp.Client {
			called = true
			return nil
		})
		assert.NotNil(t, newClientForTesting)
		newClientForTesting(nil)
		assert.True(t, called)
		newClientForTesting = nil // Reset for other tests
	})

	t.Run("SetConnectForTesting", func(t *testing.T) {
		var called bool
		SetConnectForTesting(func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
			called = true
			return nil, nil
		})
		assert.NotNil(t, connectForTesting)
		connectForTesting(nil, context.Background(), nil, nil)
		assert.True(t, called)
		connectForTesting = nil // Reset for other tests
	})
}

func TestMcpPrompt_Service(t *testing.T) {
	prompt := &mcpPrompt{service: "test-service"}
	assert.Equal(t, "test-service", prompt.Service())
}

func TestMcpResource_Service(t *testing.T) {
	resource := &mcpResource{service: "test-service"}
	assert.Equal(t, "test-service", resource.Service())
}

func TestMcpResource_Subscribe(t *testing.T) {
	resource := &mcpResource{}
	err := resource.Subscribe(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not yet implemented")
}

func TestMcpPrompt_Get(t *testing.T) {
	conn := &mcpConnection{
		stdioConfig: &configv1.McpStdioConnection{},
	}
	p := &mcpPrompt{
		mcpPrompt: &mcp.Prompt{Name: "test-prompt"},
		service:   "test-service",
		mcpConnection: conn,
	}

	ctx := context.Background()

	originalConnect := connectForTesting
	SetConnectForTesting(func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
		mockSession := new(MockClientSession)
		mockSession.On("GetPrompt", ctx, mock.Anything).Return(&mcp.GetPromptResult{
			Description: "desc",
		}, nil)
		mockSession.On("Close").Return(nil)
		return mockSession, nil
	})
	defer func() { connectForTesting = originalConnect }()

	res, err := p.Get(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, "desc", res.Description)
}

func TestMcpResource_Read(t *testing.T) {
	conn := &mcpConnection{
		stdioConfig: &configv1.McpStdioConnection{},
	}
	r := &mcpResource{
		mcpResource: &mcp.Resource{URI: "uri1"},
		service:     "test-service",
		mcpConnection: conn,
	}

	ctx := context.Background()

	originalConnect := connectForTesting
	SetConnectForTesting(func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
		mockSession := new(MockClientSession)
		mockSession.On("ReadResource", ctx, mock.Anything).Return(&mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{},
		}, nil)
		mockSession.On("Close").Return(nil)
		return mockSession, nil
	})
	defer func() { connectForTesting = originalConnect }()

	res, err := r.Read(ctx)
	require.NoError(t, err)
	assert.NotNil(t, res)
}

func TestMCPUpstream_createAndRegisterMCPItemsFromStdio(t *testing.T) {
	tm := newMockToolManager()
	pm := newMockPromptManager()
	rm := newMockResourceManager()

	stdio := &configv1.McpStdioConnection{}
	stdio.SetCommand("echo")
	mcpService := &configv1.McpUpstreamService{}
	mcpService.SetStdioConnection(stdio)
	mcpService.SetToolAutoDiscovery(true)
	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetMcpService(mcpService)

	tm.AddServiceInfo("test-service", &tool.ServiceInfo{Config: serviceConfig})

	upstream := &MCPUpstream{}

	originalConnect := connectForTesting
	SetConnectForTesting(func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
		mockSession := new(MockClientSession)
		mockSession.On("ListTools", ctx, mock.Anything).Return(&mcp.ListToolsResult{
			Tools: []*mcp.Tool{
				{Name: "tool1", Description: "desc1"},
			},
		}, nil)
		mockSession.On("ListPrompts", ctx, mock.Anything).Return(&mcp.ListPromptsResult{
			Prompts: []*mcp.Prompt{
				{Name: "prompt1", Description: "desc1"},
			},
		}, nil)
		mockSession.On("ListResources", ctx, mock.Anything).Return(&mcp.ListResourcesResult{
			Resources: []*mcp.Resource{
				{Name: "resource1", URI: "uri1"},
			},
		}, nil)
		mockSession.On("Close").Return(nil)
		return mockSession, nil
	})
	defer func() { connectForTesting = originalConnect }()

	tools, resources, err := upstream.createAndRegisterMCPItemsFromStdio(context.Background(), "test-service", stdio, tm, pm, rm, false, serviceConfig)
	require.NoError(t, err)
	assert.Len(t, tools, 1)
	assert.Equal(t, "tool1", tools[0].GetName())
	assert.Len(t, resources, 1)
	assert.Equal(t, "resource1", resources[0].GetName())

	// Check if added to managers
	_, ok := tm.GetTool("tool1")
	assert.True(t, ok)
	_, ok = pm.GetPrompt("prompt1")
	assert.True(t, ok)
}

func TestMCPUpstream_createAndRegisterMCPItemsFromStreamableHTTP(t *testing.T) {
	tm := newMockToolManager()
	pm := newMockPromptManager()
	rm := newMockResourceManager()

	httpConn := &configv1.McpStreamableHttpConnection{}
	httpConn.SetHttpAddress("http://localhost:8080")
	mcpService := &configv1.McpUpstreamService{}
	mcpService.SetHttpConnection(httpConn)
	mcpService.SetToolAutoDiscovery(true)
	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetMcpService(mcpService)

	tm.AddServiceInfo("test-service", &tool.ServiceInfo{Config: serviceConfig})

	upstream := &MCPUpstream{}

	originalConnect := connectForTesting
	SetConnectForTesting(func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
		mockSession := new(MockClientSession)
		mockSession.On("ListTools", ctx, mock.Anything).Return(&mcp.ListToolsResult{
			Tools: []*mcp.Tool{
				{Name: "tool1", Description: "desc1"},
			},
		}, nil)
		mockSession.On("ListPrompts", ctx, mock.Anything).Return(&mcp.ListPromptsResult{
			Prompts: []*mcp.Prompt{
				{Name: "prompt1", Description: "desc1"},
			},
		}, nil)
		mockSession.On("ListResources", ctx, mock.Anything).Return(&mcp.ListResourcesResult{
			Resources: []*mcp.Resource{
				{Name: "resource1", URI: "uri1"},
			},
		}, nil)
		mockSession.On("Close").Return(nil)
		return mockSession, nil
	})
	defer func() { connectForTesting = originalConnect }()

	tools, resources, err := upstream.createAndRegisterMCPItemsFromStreamableHTTP(context.Background(), "test-service", httpConn, tm, pm, rm, false, serviceConfig)
	require.NoError(t, err)
	assert.Len(t, tools, 1)
	assert.Len(t, resources, 1)
}

func TestMCPUpstream_DisabledItems(t *testing.T) {
	tm := newMockToolManager()
	pm := newMockPromptManager()
	rm := newMockResourceManager()

	stdio := &configv1.McpStdioConnection{}
	stdio.SetCommand("echo")
	mcpService := &configv1.McpUpstreamService{}
	mcpService.SetStdioConnection(stdio)
	mcpService.SetToolAutoDiscovery(true)

	// Disabled Tool Config
	toolDef := configv1.ToolDefinition_builder{
		Name:    proto.String("tool1"),
		CallId:  proto.String("call1"),
		Disable: proto.Bool(true),
	}.Build()
	mcpService.SetTools([]*configv1.ToolDefinition{toolDef})

	// Disabled Prompt Config
	promptDef := configv1.PromptDefinition_builder{
		Name:    proto.String("prompt1"),
		Disable: proto.Bool(true),
	}.Build()
	mcpService.SetPrompts([]*configv1.PromptDefinition{promptDef})

	// Disabled Resource Config
	resourceDef := configv1.ResourceDefinition_builder{
		Name:    proto.String("resource1"),
		Disable: proto.Bool(true),
	}.Build()
	mcpService.SetResources([]*configv1.ResourceDefinition{resourceDef})

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetMcpService(mcpService)

	tm.AddServiceInfo("test-service", &tool.ServiceInfo{Config: serviceConfig})

	upstream := NewMCPUpstream().(*MCPUpstream)

	originalConnect := connectForTesting
	SetConnectForTesting(func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
		mockSession := new(MockClientSession)
		// Return items that match disabled config
		mockSession.On("ListTools", ctx, mock.Anything).Return(&mcp.ListToolsResult{
			Tools: []*mcp.Tool{
				{Name: "tool1", Description: "desc1"},
			},
		}, nil)
		mockSession.On("ListPrompts", ctx, mock.Anything).Return(&mcp.ListPromptsResult{
			Prompts: []*mcp.Prompt{
				{Name: "prompt1", Description: "desc1"},
			},
		}, nil)
		mockSession.On("ListResources", ctx, mock.Anything).Return(&mcp.ListResourcesResult{
			Resources: []*mcp.Resource{
				{Name: "resource1", URI: "uri1"},
			},
		}, nil)
		mockSession.On("Close").Return(nil)
		return mockSession, nil
	})
	defer func() { connectForTesting = originalConnect }()

	_, tools, resources, err := upstream.Register(context.Background(), serviceConfig, tm, pm, rm, false)
	require.NoError(t, err)
	assert.Empty(t, tools)
	assert.Empty(t, resources)

	// Verify nothing added to managers
	_, ok := tm.GetTool("tool1")
	assert.False(t, ok)
	_, ok = pm.GetPrompt("prompt1")
	assert.False(t, ok)
}
