/*
 * Copyright 2025 Author(s) of MCP-XY
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

package websocket

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/mcpxy/core/pkg/bus"
	"github.com/mcpxy/core/pkg/client"
	"github.com/mcpxy/core/pkg/pool"
	"github.com/mcpxy/core/pkg/prompt"
	"github.com/mcpxy/core/pkg/resource"
	"github.com/mcpxy/core/pkg/tool"
	"github.com/mcpxy/core/pkg/util"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// MockToolManager is a mock implementation of the ToolManagerInterface.
type MockToolManager struct {
	mu      sync.Mutex
	tools   map[string]tool.Tool
	lastErr error
}

func NewMockToolManager(bus *bus.BusProvider) *MockToolManager {
	return &MockToolManager{
		tools: make(map[string]tool.Tool),
	}
}

func (m *MockToolManager) AddTool(t tool.Tool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.lastErr != nil {
		return m.lastErr
	}
	toolID, _ := util.GenerateToolID(t.Tool().GetServiceId(), t.Tool().GetName())
	m.tools[toolID] = t
	return nil
}

func (m *MockToolManager) GetTool(name string) (tool.Tool, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tools[name]
	return t, ok
}

func (m *MockToolManager) ListTools() []tool.Tool {
	m.mu.Lock()
	defer m.mu.Unlock()
	tools := make([]tool.Tool, 0, len(m.tools))
	for _, t := range m.tools {
		tools = append(tools, t)
	}
	return tools
}

func (m *MockToolManager) ClearToolsForService(serviceKey string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, t := range m.tools {
		if t.Tool().GetServiceId() == serviceKey {
			delete(m.tools, name)
		}
	}
}

func (m *MockToolManager) SetMCPServer(provider tool.MCPServerProvider) {}

func (m *MockToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {}

func (m *MockToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	return nil, false
}

func (m *MockToolManager) ExecuteTool(ctx context.Context, req *tool.ExecutionRequest) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func TestNewWebsocketUpstream(t *testing.T) {
	poolManager := pool.NewManager()
	upstream := NewWebsocketUpstream(poolManager)
	require.NotNil(t, upstream)
	assert.IsType(t, &WebsocketUpstream{}, upstream)
}

func TestWebsocketUpstream_Register_Mocked(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		toolManager := NewMockToolManager(nil)
		poolManager := pool.NewManager()
		var promptManager prompt.PromptManagerInterface
		var resourceManager resource.ResourceManagerInterface

		upstream := NewWebsocketUpstream(poolManager)

		callDef := configv1.WebsocketCallDefinition_builder{
			Schema: configv1.ToolSchema_builder{
				Name:        proto.String("echo"),
				Description: proto.String("Echoes a message"),
			}.Build(),
		}.Build()

		websocketService := &configv1.WebsocketUpstreamService{}
		websocketService.SetAddress("ws://localhost:8080/echo")
		websocketService.SetCalls([]*configv1.WebsocketCallDefinition{callDef})

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-websocket-service")
		serviceConfig.SetWebsocketService(websocketService)

		serviceKey, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)

		tools := toolManager.ListTools()
		assert.Len(t, tools, 1)

		toolID, _ := util.GenerateToolID(serviceKey, "echo")
		_, ok := toolManager.GetTool(toolID)
		assert.True(t, ok, "tool should be registered")
	})

	t.Run("nil service config", func(t *testing.T) {
		toolManager := NewMockToolManager(nil)
		poolManager := pool.NewManager()
		var promptManager prompt.PromptManagerInterface
		var resourceManager resource.ResourceManagerInterface
		upstream := NewWebsocketUpstream(poolManager)

		_, _, err := upstream.Register(context.Background(), nil, toolManager, promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "service config is nil")
	})

	t.Run("nil websocket service config", func(t *testing.T) {
		toolManager := NewMockToolManager(nil)
		poolManager := pool.NewManager()
		var promptManager prompt.PromptManagerInterface
		var resourceManager resource.ResourceManagerInterface
		upstream := NewWebsocketUpstream(poolManager)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-websocket-service")
		serviceConfig.SetWebsocketService(nil)

		_, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Equal(t, "websocket service config is nil", err.Error())
	})

	t.Run("add tool error", func(t *testing.T) {
		toolManager := NewMockToolManager(nil)
		toolManager.lastErr = errors.New("failed to add tool")
		poolManager := pool.NewManager()
		var promptManager prompt.PromptManagerInterface
		var resourceManager resource.ResourceManagerInterface
		upstream := NewWebsocketUpstream(poolManager)

		callDef := configv1.WebsocketCallDefinition_builder{
			Schema: configv1.ToolSchema_builder{
				Name: proto.String("echo"),
			}.Build(),
		}.Build()

		websocketService := &configv1.WebsocketUpstreamService{}
		websocketService.SetAddress("ws://localhost:8080/echo")
		websocketService.SetCalls([]*configv1.WebsocketCallDefinition{callDef})

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-websocket-service")
		serviceConfig.SetWebsocketService(websocketService)

		_, discoveredTools, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)
		assert.Empty(t, discoveredTools)
	})

	t.Run("authenticator creation fails", func(t *testing.T) {
		toolManager := NewMockToolManager(nil)
		poolManager := pool.NewManager()
		var promptManager prompt.PromptManagerInterface
		var resourceManager resource.ResourceManagerInterface
		upstream := NewWebsocketUpstream(poolManager)

		callDef := configv1.WebsocketCallDefinition_builder{
			Schema: configv1.ToolSchema_builder{
				Name: proto.String("echo"),
			}.Build(),
		}.Build()

		websocketService := &configv1.WebsocketUpstreamService{}
		websocketService.SetAddress("ws://localhost:8080/echo")
		websocketService.SetCalls([]*configv1.WebsocketCallDefinition{callDef})

		authConfig := (&configv1.UpstreamAuthentication_builder{
			ApiKey: &configv1.UpstreamAPIKeyAuth{},
		}).Build()

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("auth-fail-service")
		serviceConfig.SetWebsocketService(websocketService)
		serviceConfig.SetUpstreamAuthentication(authConfig)

		_, discoveredTools, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)
		assert.Len(t, discoveredTools, 1, "a tool should be discovered if auth config is incomplete")
	})

	t.Run("tool registration with fallback operation ID", func(t *testing.T) {
		toolManager := NewMockToolManager(nil)
		poolManager := pool.NewManager()
		var promptManager prompt.PromptManagerInterface
		var resourceManager resource.ResourceManagerInterface
		upstream := NewWebsocketUpstream(poolManager)

		// Fallback to description
		callDef1 := configv1.WebsocketCallDefinition_builder{
			Schema: configv1.ToolSchema_builder{
				Description: proto.String("This is a test description"),
			}.Build(),
		}.Build()

		callDef2 := configv1.WebsocketCallDefinition_builder{
			Schema: configv1.ToolSchema_builder{
				Description: proto.String(""),
			}.Build(),
		}.Build()

		websocketService := &configv1.WebsocketUpstreamService{}
		websocketService.SetAddress("ws://localhost:8080/echo")
		websocketService.SetCalls([]*configv1.WebsocketCallDefinition{callDef1, callDef2})

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-service-fallback")
		serviceConfig.SetWebsocketService(websocketService)

		serviceKey, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)

		tools := toolManager.ListTools()
		assert.Len(t, tools, 2)

		sanitizedName := util.SanitizeOperationID("This is a test description")
		toolID1, _ := util.GenerateToolID(serviceKey, sanitizedName)
		_, ok := toolManager.GetTool(toolID1)
		assert.True(t, ok, "Tool with sanitized description should be found, expected %s", toolID1)

		toolID2, _ := util.GenerateToolID(serviceKey, "op1")
		_, ok = toolManager.GetTool(toolID2)
		assert.True(t, ok, "tool should be registered with op index")
	})

	t.Run("correct input schema generation", func(t *testing.T) {
		toolManager := NewMockToolManager(nil)
		poolManager := pool.NewManager()
		upstream := NewWebsocketUpstream(poolManager)

		param1 := configv1.WebsocketParameterMapping_builder{
			Schema: configv1.ParameterSchema_builder{
				Name: proto.String("param1"),
			}.Build(),
		}.Build()
		param2 := configv1.WebsocketParameterMapping_builder{
			Schema: configv1.ParameterSchema_builder{
				Name: proto.String("param2"),
			}.Build(),
		}.Build()

		callDef := configv1.WebsocketCallDefinition_builder{
			Schema: configv1.ToolSchema_builder{
				Name: proto.String("test-tool"),
			}.Build(),
			Parameters: []*configv1.WebsocketParameterMapping{param1, param2},
		}.Build()

		websocketService := configv1.WebsocketUpstreamService_builder{
			Address: proto.String("ws://localhost:8080/test"),
			Calls:   []*configv1.WebsocketCallDefinition{callDef},
		}.Build()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name:             proto.String("test-websocket-service"),
			WebsocketService: websocketService,
		}.Build()

		serviceKey, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, nil, nil, false)
		require.NoError(t, err)

		toolID, _ := util.GenerateToolID(serviceKey, "test-tool")
		registeredTool, ok := toolManager.GetTool(toolID)
		require.True(t, ok)

		inputSchema := registeredTool.Tool().GetAnnotations().GetInputSchema()
		require.NotNil(t, inputSchema)
		assert.Equal(t, "object", inputSchema.GetFields()["type"].GetStringValue())

		properties := inputSchema.GetFields()["properties"].GetStructValue().GetFields()
		assert.Contains(t, properties, "param1")
		assert.Contains(t, properties, "param2")
	})
}

func TestWebsocketUpstream_Register_Integration(t *testing.T) {
	upgrader := websocket.Upgrader{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	poolManager := pool.NewManager()
	tm := tool.NewToolManager(nil)

	t.Run("successful registration", func(t *testing.T) {
		upstream := NewWebsocketUpstream(poolManager)

		apiKeyAuth := &configv1.UpstreamAPIKeyAuth{}
		apiKeyAuth.SetHeaderName("X-API-Key")
		apiKeyAuth.SetApiKey("test-key")

		authConfig := &configv1.UpstreamAuthentication{}
		authConfig.SetApiKey(apiKeyAuth)

		call1 := configv1.WebsocketCallDefinition_builder{
			Schema: configv1.ToolSchema_builder{
				Name:        proto.String("test-op"),
				Description: proto.String("A test operation"),
			}.Build(),
		}.Build()

		call2 := configv1.WebsocketCallDefinition_builder{
			Schema: configv1.ToolSchema_builder{
				Description: proto.String("Another test operation"),
			}.Build(),
		}.Build()

		wsService := &configv1.WebsocketUpstreamService{}
		wsService.SetAddress(wsURL)
		wsService.SetCalls([]*configv1.WebsocketCallDefinition{call1, call2})

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-service")
		serviceConfig.SetUpstreamAuthentication(authConfig)
		serviceConfig.SetWebsocketService(wsService)

		serviceKey, discoveredTools, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		require.NoError(t, err)
		expectedKey, _ := util.GenerateID("test-service")
		assert.Equal(t, expectedKey, serviceKey)
		assert.Len(t, discoveredTools, 2)
		_, ok := pool.Get[*client.WebsocketClientWrapper](poolManager, serviceKey)
		assert.True(t, ok)
	})

	t.Run("nil websocket service config", func(t *testing.T) {
		upstream := NewWebsocketUpstream(poolManager)
		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("nil-config-service")

		_, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "websocket service config is nil")
	})

	t.Run("invalid service name", func(t *testing.T) {
		upstream := NewWebsocketUpstream(poolManager)
		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("")
		serviceConfig.SetWebsocketService(&configv1.WebsocketUpstreamService{})

		_, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name cannot be empty")
	})

	t.Run("authenticator creation fails", func(t *testing.T) {
		upstream := NewWebsocketUpstream(poolManager)

		call1 := configv1.WebsocketCallDefinition_builder{
			Schema: configv1.ToolSchema_builder{
				Name: proto.String("test-op"),
			}.Build(),
		}.Build()

		wsService := &configv1.WebsocketUpstreamService{}
		wsService.SetAddress(wsURL)
		wsService.SetCalls([]*configv1.WebsocketCallDefinition{call1})

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("auth-fail-service")
		// Intentionally not setting auth method on UpstreamAuthentication, which is a valid scenario (no auth).
		serviceConfig.SetUpstreamAuthentication(&configv1.UpstreamAuthentication{})
		serviceConfig.SetWebsocketService(wsService)

		serviceKey, tools, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		require.NoError(t, err)
		expectedKey, _ := util.GenerateID("auth-fail-service")
		assert.Equal(t, expectedKey, serviceKey)
		assert.Len(t, tools, 1, "expected one tool when authenticator is not configured")
	})

	t.Run("tool registration with fallback operation ID", func(t *testing.T) {
		tm := tool.NewToolManager(nil)
		upstream := NewWebsocketUpstream(poolManager)

		call1 := configv1.WebsocketCallDefinition_builder{
			Schema: configv1.ToolSchema_builder{
				Description: proto.String("A test operation"),
			}.Build(),
		}.Build()

		call2 := configv1.WebsocketCallDefinition_builder{
			Schema: configv1.ToolSchema_builder{
				Description: proto.String("Another test operation"),
			}.Build(),
		}.Build()

		wsService := &configv1.WebsocketUpstreamService{}
		wsService.SetAddress(wsURL)
		wsService.SetCalls([]*configv1.WebsocketCallDefinition{call1, call2})

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("fallback-op-id")
		serviceConfig.SetWebsocketService(wsService)

		serviceKey, discoveredTools, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		require.NoError(t, err)
		expectedKey, _ := util.GenerateID("fallback-op-id")
		assert.Equal(t, expectedKey, serviceKey)
		assert.Len(t, discoveredTools, 2)
	})
}

func TestWebsocketUpstream_Register_WithReload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer server.Close()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	poolManager := pool.NewManager()
	tm := tool.NewToolManager(nil)
	upstream := NewWebsocketUpstream(poolManager)

	call1 := configv1.WebsocketCallDefinition_builder{
		Schema: configv1.ToolSchema_builder{
			Name: proto.String("test-op"),
		}.Build(),
	}.Build()

	wsService := &configv1.WebsocketUpstreamService{}
	wsService.SetAddress(wsURL)
	wsService.SetCalls([]*configv1.WebsocketCallDefinition{call1})

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("reload-test")
	serviceConfig.SetWebsocketService(wsService)

	serviceKey, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)
	toolID, _ := util.GenerateToolID(serviceKey, "test-op")
	retrievedTool, ok := tm.GetTool(toolID)
	assert.True(t, ok)
	assert.NotNil(t, retrievedTool)

	_, _, err = upstream.Register(context.Background(), serviceConfig, tm, nil, nil, true)
	require.NoError(t, err)
	retrievedTool, ok = tm.GetTool(toolID)
	assert.True(t, ok)
	assert.NotNil(t, retrievedTool)
}
