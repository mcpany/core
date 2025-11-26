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

package webrtc

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
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

func NewMockToolManager() *MockToolManager {
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
	sanitizedToolName, _ := util.SanitizeToolName(t.Tool().GetName())
	toolID := t.Tool().GetServiceId() + "." + sanitizedToolName
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

func (m *MockToolManager) ClearToolsForService(serviceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, t := range m.tools {
		if t.Tool().GetServiceId() == serviceID {
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

func (m *MockToolManager) AddMiddleware(middleware tool.ToolExecutionMiddleware) {
}

// MockPromptManager is a mock implementation of the PromptManagerInterface.
type MockPromptManager struct {
	mu      sync.Mutex
	prompts map[string]prompt.Prompt
}

func NewMockPromptManager() *MockPromptManager {
	return &MockPromptManager{
		prompts: make(map[string]prompt.Prompt),
	}
}

func (m *MockPromptManager) AddPrompt(p prompt.Prompt) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.prompts[p.Prompt().Name] = p
}

func (m *MockPromptManager) UpdatePrompt(p prompt.Prompt) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.prompts[p.Prompt().Name] = p
}

func (m *MockPromptManager) GetPrompt(name string) (prompt.Prompt, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.prompts[name]
	return p, ok
}

func (m *MockPromptManager) ListPrompts() []prompt.Prompt {
	m.mu.Lock()
	defer m.mu.Unlock()
	prompts := make([]prompt.Prompt, 0, len(m.prompts))
	for _, p := range m.prompts {
		prompts = append(prompts, p)
	}
	return prompts
}

func (m *MockPromptManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	return nil, false
}

func (m *MockPromptManager) SetMCPServer(mcpServer prompt.MCPServerProvider) {
}

func (m *MockPromptManager) ClearPromptsForService(serviceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, p := range m.prompts {
		if p.Service() == serviceID {
			delete(m.prompts, name)
		}
	}
}

// MockResourceManager is a mock implementation of the ResourceManagerInterface.
type MockResourceManager struct {
	mu        sync.Mutex
	resources map[string]resource.Resource
}

func NewMockResourceManager() *MockResourceManager {
	return &MockResourceManager{
		resources: make(map[string]resource.Resource),
	}
}

func (m *MockResourceManager) AddResource(r resource.Resource) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.resources[r.Resource().Name] = r
}

func (m *MockResourceManager) GetResource(name string) (resource.Resource, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.resources[name]
	return r, ok
}

func (m* MockResourceManager) RemoveResource(uri string) {
}

func (m *MockResourceManager) ListResources() []resource.Resource {
	m.mu.Lock()
	defer m.mu.Unlock()
	resources := make([]resource.Resource, 0, len(m.resources))
	for _, r := range m.resources {
		resources = append(resources, r)
	}
	return resources
}

func (m *MockResourceManager) OnListChanged(f func()) {
}

func (m *MockResourceManager) ClearResourcesForService(serviceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, r := range m.resources {
		if r.Service() == serviceID {
			delete(m.resources, name)
		}
	}
}

func TestNewWebrtcUpstream(t *testing.T) {
	poolManager := pool.NewManager()
	upstream := NewWebrtcUpstream(poolManager)
	require.NotNil(t, upstream)
	assert.IsType(t, &WebrtcUpstream{}, upstream)
}

func TestWebrtcUpstream_Register(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		toolManager := NewMockToolManager()
		poolManager := pool.NewManager()
		var promptManager prompt.PromptManagerInterface
		var resourceManager resource.ResourceManagerInterface

		upstream := NewWebrtcUpstream(poolManager)

		toolDef := configv1.ToolDefinition_builder{
			Name:        proto.String("echo"),
			Description: proto.String("Echoes a message"),
			CallId:      proto.String("echo-call"),
		}.Build()

		webrtcService := &configv1.WebrtcUpstreamService{}
		webrtcService.SetAddress("http://localhost:8080/signal")
		webrtcService.SetTools([]*configv1.ToolDefinition{toolDef})
		calls := make(map[string]*configv1.WebrtcCallDefinition)
		calls["echo-call"] = configv1.WebrtcCallDefinition_builder{
			Id: proto.String("echo-call"),
		}.Build()
		webrtcService.SetCalls(calls)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-webrtc-service")
		serviceConfig.SetWebrtcService(webrtcService)

		serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)

		tools := toolManager.ListTools()
		assert.Len(t, tools, 1)

		sanitizedToolName, _ := util.SanitizeToolName("echo")
		toolID := serviceID + "." + sanitizedToolName
		_, ok := toolManager.GetTool(toolID)
		assert.True(t, ok, "tool should be registered")
	})

	t.Run("nil service config", func(t *testing.T) {
		toolManager := NewMockToolManager()
		poolManager := pool.NewManager()
		var promptManager prompt.PromptManagerInterface
		var resourceManager resource.ResourceManagerInterface
		upstream := NewWebrtcUpstream(poolManager)

		_, _, _, err := upstream.Register(context.Background(), nil, toolManager, promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "service config is nil")
	})

	t.Run("nil webrtc service config", func(t *testing.T) {
		toolManager := NewMockToolManager()
		poolManager := pool.NewManager()
		var promptManager prompt.PromptManagerInterface
		var resourceManager resource.ResourceManagerInterface
		upstream := NewWebrtcUpstream(poolManager)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-webrtc-service")
		serviceConfig.SetWebrtcService(nil)

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Equal(t, "webrtc service config is nil", err.Error())
	})

	t.Run("add tool error", func(t *testing.T) {
		toolManager := NewMockToolManager()
		toolManager.lastErr = errors.New("failed to add tool")
		poolManager := pool.NewManager()
		var promptManager prompt.PromptManagerInterface
		var resourceManager resource.ResourceManagerInterface
		upstream := NewWebrtcUpstream(poolManager)

		toolDef := configv1.ToolDefinition_builder{
			Name:   proto.String("echo"),
			CallId: proto.String("echo-call"),
		}.Build()

		webrtcService := &configv1.WebrtcUpstreamService{}
		webrtcService.SetAddress("http://localhost:8080/signal")
		webrtcService.SetTools([]*configv1.ToolDefinition{toolDef})
		calls := make(map[string]*configv1.WebrtcCallDefinition)
		calls["echo-call"] = configv1.WebrtcCallDefinition_builder{
			Id: proto.String("echo-call"),
		}.Build()
		webrtcService.SetCalls(calls)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-webrtc-service")
		serviceConfig.SetWebrtcService(webrtcService)

		_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)
		assert.Empty(t, discoveredTools)
	})

	t.Run("authenticator error", func(t *testing.T) {
		toolManager := NewMockToolManager()
		poolManager := pool.NewManager()
		var promptManager prompt.PromptManagerInterface
		var resourceManager resource.ResourceManagerInterface
		upstream := NewWebrtcUpstream(poolManager)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-webrtc-service")
		authConfig := &configv1.UpstreamAuthentication{}
		apiKeyAuth := &configv1.UpstreamAPIKeyAuth{}
		apiKeyAuth.SetHeaderName("") // Invalid header name
		authConfig.SetApiKey(apiKeyAuth)
		serviceConfig.SetUpstreamAuthentication(authConfig)

		webrtcService := &configv1.WebrtcUpstreamService{}
		webrtcService.SetAddress("http://localhost:8080/signal")
		toolDef := &configv1.ToolDefinition{}
		toolDef.SetName("echo")
		toolDef.SetCallId("echo-call")
		webrtcService.SetTools([]*configv1.ToolDefinition{toolDef})
		callDef := &configv1.WebrtcCallDefinition{}
		callDef.SetId("echo-call")
		webrtcService.SetCalls(map[string]*configv1.WebrtcCallDefinition{
			"echo-call": callDef,
		})
		serviceConfig.SetWebrtcService(webrtcService)

		_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)
		assert.Empty(t, discoveredTools)
		assert.Empty(t, toolManager.ListTools())
	})

	t.Run("missing call id", func(t *testing.T) {
		toolManager := NewMockToolManager()
		poolManager := pool.NewManager()
		var promptManager prompt.PromptManagerInterface
		var resourceManager resource.ResourceManagerInterface
		upstream := NewWebrtcUpstream(poolManager)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-webrtc-service")
		webrtcService := &configv1.WebrtcUpstreamService{}
		webrtcService.SetAddress("http://localhost:8080/signal")
		toolDef := &configv1.ToolDefinition{}
		toolDef.SetName("echo")
		toolDef.SetCallId("non-existent-call-id")
		webrtcService.SetTools([]*configv1.ToolDefinition{toolDef})
		serviceConfig.SetWebrtcService(webrtcService)

		_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)
		assert.Empty(t, discoveredTools)
		assert.Empty(t, toolManager.ListTools())
	})

	t.Run("successful prompt and resource registration", func(t *testing.T) {
		toolManager := NewMockToolManager()
		poolManager := pool.NewManager()
		promptManager := NewMockPromptManager()
		resourceManager := NewMockResourceManager()
		upstream := NewWebrtcUpstream(poolManager)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-webrtc-service-with-prompts-and-resources")

		webrtcService := &configv1.WebrtcUpstreamService{}
		webrtcService.SetAddress("http://localhost:8080/signal")
		toolDef := &configv1.ToolDefinition{}
		toolDef.SetName("get-weather")
		toolDef.SetCallId("get-weather-call")
		webrtcService.SetTools([]*configv1.ToolDefinition{toolDef})

		callDef := &configv1.WebrtcCallDefinition{}
		callDef.SetId("get-weather-call")
		webrtcService.SetCalls(map[string]*configv1.WebrtcCallDefinition{
			"get-weather-call": callDef,
		})

		promptDef := &configv1.PromptDefinition{}
		promptDef.SetName("weather-prompt")
		promptMessage := &configv1.PromptMessage{}
		textContent := &configv1.TextContent{}
		textContent.SetText("What is the weather in {{.location}}?")
		promptMessage.SetText(textContent)
		promptDef.SetMessages([]*configv1.PromptMessage{promptMessage})
		webrtcService.SetPrompts([]*configv1.PromptDefinition{promptDef})

		resourceDef := &configv1.ResourceDefinition{}
		resourceDef.SetName("weather-resource")
		dynamicResource := &configv1.DynamicResource{}
		webrtcCall := &configv1.WebrtcCallDefinition{}
		webrtcCall.SetId("get-weather-call")
		dynamicResource.SetWebrtcCall(webrtcCall)
		resourceDef.SetDynamic(dynamicResource)
		webrtcService.SetResources([]*configv1.ResourceDefinition{resourceDef})

		serviceConfig.SetWebrtcService(webrtcService)

		serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)

		_, ok := promptManager.GetPrompt(serviceID + ".weather-prompt")
		assert.True(t, ok, "prompt should be registered")

		_, ok = resourceManager.GetResource("weather-resource")
		assert.True(t, ok, "resource should be registered")
	})

	t.Run("sanitizer failure", func(t *testing.T) {
		toolManager := NewMockToolManager()
		poolManager := pool.NewManager()
		promptManager := NewMockPromptManager()
		resourceManager := NewMockResourceManager()
		upstream := NewWebrtcUpstream(poolManager).(*WebrtcUpstream)
		upstream.toolNameSanitizer = func(name string) (string, error) {
			return "", errors.New("sanitization failed")
		}

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-webrtc-service-with-sanitizer-failure")

		webrtcService := &configv1.WebrtcUpstreamService{}
		webrtcService.SetAddress("http://localhost:8080/signal")
		toolDef := &configv1.ToolDefinition{}
		toolDef.SetName("get-weather")
		toolDef.SetCallId("get-weather-call")
		webrtcService.SetTools([]*configv1.ToolDefinition{toolDef})

		callDef := &configv1.WebrtcCallDefinition{}
		callDef.SetId("get-weather-call")
		webrtcService.SetCalls(map[string]*configv1.WebrtcCallDefinition{
			"get-weather-call": callDef,
		})

		resourceDef := &configv1.ResourceDefinition{}
		resourceDef.SetName("weather-resource")
		dynamicResource := &configv1.DynamicResource{}
		webrtcCall := &configv1.WebrtcCallDefinition{}
		webrtcCall.SetId("get-weather-call")
		dynamicResource.SetWebrtcCall(webrtcCall)
		resourceDef.SetDynamic(dynamicResource)
		webrtcService.SetResources([]*configv1.ResourceDefinition{resourceDef})

		serviceConfig.SetWebrtcService(webrtcService)

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)

		_, ok := resourceManager.GetResource("weather-resource")
		assert.False(t, ok, "resource should not be registered")
	})
}

func TestWebrtcUpstream_Register_ToolNameGeneration(t *testing.T) {
	toolManager := NewMockToolManager()
	poolManager := pool.NewManager()
	var promptManager prompt.PromptManagerInterface
	var resourceManager resource.ResourceManagerInterface
	upstream := NewWebrtcUpstream(poolManager)

	toolDef := configv1.ToolDefinition_builder{
		Description: proto.String("A test description"),
		CallId:      proto.String("test-call"),
	}.Build()

	webrtcService := &configv1.WebrtcUpstreamService{}
	webrtcService.SetAddress("http://localhost:8080/signal")
	webrtcService.SetTools([]*configv1.ToolDefinition{toolDef})
	calls := make(map[string]*configv1.WebrtcCallDefinition)
	calls["test-call"] = configv1.WebrtcCallDefinition_builder{
		Id: proto.String("test-call"),
	}.Build()
	webrtcService.SetCalls(calls)

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-webrtc-service-tool-name-generation")
	serviceConfig.SetWebrtcService(webrtcService)

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
	require.NoError(t, err)

	tools := toolManager.ListTools()
	assert.Len(t, tools, 1)
	assert.Equal(t, util.SanitizeOperationID("A test description"), tools[0].Tool().GetName())
}
