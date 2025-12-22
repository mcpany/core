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

func (m *MockToolManager) SetMCPServer(_ tool.MCPServerProvider) {}

func (m *MockToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}

func (m *MockToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}

func (m *MockToolManager) GetServiceInfo(_ string) (*tool.ServiceInfo, bool) {
	return nil, false
}

func (m *MockToolManager) ListServices() []*tool.ServiceInfo {
	return nil
}

func (m *MockToolManager) ExecuteTool(_ context.Context, _ *tool.ExecutionRequest) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *MockToolManager) AddMiddleware(_ tool.ExecutionMiddleware) {
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

func (m *MockPromptManager) GetServiceInfo(_ string) (*tool.ServiceInfo, bool) {
	return nil, false
}

func (m *MockPromptManager) SetMCPServer(_ prompt.MCPServerProvider) {
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

func (m *MockResourceManager) RemoveResource(_ string) {
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

func (m *MockResourceManager) OnListChanged(_ func()) {
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

func TestUpstream(t *testing.T) {
	poolManager := pool.NewManager()
	upstream := NewUpstream(poolManager)
	require.NotNil(t, upstream)
	assert.IsType(t, &Upstream{}, upstream)
}

func TestUpstream_Register(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		toolManager := NewMockToolManager()
		poolManager := pool.NewManager()
		var promptManager prompt.ManagerInterface
		var resourceManager resource.ManagerInterface

		upstream := NewUpstream(poolManager)

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
		var promptManager prompt.ManagerInterface
		var resourceManager resource.ManagerInterface
		upstream := NewUpstream(poolManager)

		_, _, _, err := upstream.Register(context.Background(), nil, toolManager, promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "service config is nil")
	})

	t.Run("nil webrtc service config", func(t *testing.T) {
		toolManager := NewMockToolManager()
		poolManager := pool.NewManager()
		var promptManager prompt.ManagerInterface
		var resourceManager resource.ManagerInterface
		upstream := NewUpstream(poolManager)

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
		var promptManager prompt.ManagerInterface
		var resourceManager resource.ManagerInterface
		upstream := NewUpstream(poolManager)

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
		var promptManager prompt.ManagerInterface
		var resourceManager resource.ManagerInterface
		upstream := NewUpstream(poolManager)

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
		var promptManager prompt.ManagerInterface
		var resourceManager resource.ManagerInterface
		upstream := NewUpstream(poolManager)

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
		upstream := NewUpstream(poolManager)

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
		upstream := NewUpstream(poolManager).(*Upstream)
		upstream.toolNameSanitizer = func(_ string) (string, error) {
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

func TestUpstream_Register_ToolNameGeneration(t *testing.T) {
	toolManager := NewMockToolManager()
	poolManager := pool.NewManager()
	var promptManager prompt.ManagerInterface
	var resourceManager resource.ManagerInterface
	upstream := NewUpstream(poolManager)

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

func TestUpstream_Register_CornerCases(t *testing.T) {
	t.Run("disabled tool", func(t *testing.T) {
		toolManager := NewMockToolManager()
		poolManager := pool.NewManager()
		upstream := NewUpstream(poolManager)

		toolDef := configv1.ToolDefinition_builder{
			Name:    proto.String("disabled-tool"),
			CallId:  proto.String("call-id"),
			Disable: proto.Bool(true),
		}.Build()

		webrtcService := &configv1.WebrtcUpstreamService{}
		webrtcService.SetAddress("http://localhost:8080/signal")
		webrtcService.SetTools([]*configv1.ToolDefinition{toolDef})

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-webrtc-disabled")
		serviceConfig.SetWebrtcService(webrtcService)

		_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, nil, nil, false)
		require.NoError(t, err)
		assert.Empty(t, discoveredTools)
	})

	t.Run("empty name fallback", func(t *testing.T) {
		toolManager := NewMockToolManager()
		poolManager := pool.NewManager()
		upstream := NewUpstream(poolManager)

		toolDef := configv1.ToolDefinition_builder{
			Name:        proto.String(""), // Empty name
			Description: proto.String(""), // Empty description -> no summary
			CallId:      proto.String("call-id"),
		}.Build()

		webrtcService := &configv1.WebrtcUpstreamService{}
		webrtcService.SetAddress("http://localhost:8080/signal")
		webrtcService.SetTools([]*configv1.ToolDefinition{toolDef})
		calls := make(map[string]*configv1.WebrtcCallDefinition)
		calls["call-id"] = configv1.WebrtcCallDefinition_builder{Id: proto.String("call-id")}.Build()
		webrtcService.SetCalls(calls)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-webrtc-empty-name")
		serviceConfig.SetWebrtcService(webrtcService)

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, nil, nil, false)
		require.NoError(t, err)

		tools := toolManager.ListTools()
		require.Len(t, tools, 1)
		assert.Equal(t, "op0", tools[0].Tool().GetName())
	})

	t.Run("disabled resource", func(t *testing.T) {
		toolManager := NewMockToolManager()
		poolManager := pool.NewManager()
		resourceManager := NewMockResourceManager()
		upstream := NewUpstream(poolManager)

		webrtcService := &configv1.WebrtcUpstreamService{}
		webrtcService.SetAddress("http://localhost:8080/signal")

		resourceDef := &configv1.ResourceDefinition{}
		resourceDef.SetName("disabled-resource")
		resourceDef.SetDisable(true)
		webrtcService.SetResources([]*configv1.ResourceDefinition{resourceDef})

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-webrtc-disabled-resource")
		serviceConfig.SetWebrtcService(webrtcService)

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, nil, resourceManager, false)
		require.NoError(t, err)
		assert.Empty(t, resourceManager.ListResources())
	})

	t.Run("dynamic resource missing call", func(t *testing.T) {
		toolManager := NewMockToolManager()
		poolManager := pool.NewManager()
		resourceManager := NewMockResourceManager()
		upstream := NewUpstream(poolManager)

		webrtcService := &configv1.WebrtcUpstreamService{}
		webrtcService.SetAddress("http://localhost:8080/signal")

		resourceDef := &configv1.ResourceDefinition{}
		resourceDef.SetName("resource-missing-call")
		dynamicResource := &configv1.DynamicResource{}
		// No WebrtcCall set
		resourceDef.SetDynamic(dynamicResource)
		webrtcService.SetResources([]*configv1.ResourceDefinition{resourceDef})

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-webrtc-missing-call")
		serviceConfig.SetWebrtcService(webrtcService)

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, nil, resourceManager, false)
		require.NoError(t, err)
		assert.Empty(t, resourceManager.ListResources())
	})

	t.Run("dynamic resource call id not found", func(t *testing.T) {
		toolManager := NewMockToolManager()
		poolManager := pool.NewManager()
		resourceManager := NewMockResourceManager()
		upstream := NewUpstream(poolManager)

		webrtcService := &configv1.WebrtcUpstreamService{}
		webrtcService.SetAddress("http://localhost:8080/signal")

		resourceDef := &configv1.ResourceDefinition{}
		resourceDef.SetName("resource-call-not-found")
		dynamicResource := &configv1.DynamicResource{}
		webrtcCall := &configv1.WebrtcCallDefinition{}
		webrtcCall.SetId("unknown-call-id")
		dynamicResource.SetWebrtcCall(webrtcCall)
		resourceDef.SetDynamic(dynamicResource)
		webrtcService.SetResources([]*configv1.ResourceDefinition{resourceDef})

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-webrtc-call-not-found")
		serviceConfig.SetWebrtcService(webrtcService)

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, nil, resourceManager, false)
		require.NoError(t, err)
		assert.Empty(t, resourceManager.ListResources())
	})

	t.Run("tool not found for dynamic resource", func(t *testing.T) {
		toolManager := NewMockToolManager()
		poolManager := pool.NewManager()
		resourceManager := NewMockResourceManager()
		upstream := NewUpstream(poolManager)

		toolManager.lastErr = errors.New("fail add tool")

		toolDef := configv1.ToolDefinition_builder{
			Name:   proto.String("tool1"),
			CallId: proto.String("call1"),
		}.Build()

		webrtcService := &configv1.WebrtcUpstreamService{}
		webrtcService.SetAddress("http://localhost:8080/signal")
		webrtcService.SetTools([]*configv1.ToolDefinition{toolDef})
		calls := make(map[string]*configv1.WebrtcCallDefinition)
		calls["call1"] = configv1.WebrtcCallDefinition_builder{Id: proto.String("call1")}.Build()
		webrtcService.SetCalls(calls)

		resourceDef := &configv1.ResourceDefinition{}
		resourceDef.SetName("resource1")
		dynamicResource := &configv1.DynamicResource{}
		webrtcCall := &configv1.WebrtcCallDefinition{}
		webrtcCall.SetId("call1")
		dynamicResource.SetWebrtcCall(webrtcCall)
		resourceDef.SetDynamic(dynamicResource)
		webrtcService.SetResources([]*configv1.ResourceDefinition{resourceDef})

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-webrtc-tool-not-found")
		serviceConfig.SetWebrtcService(webrtcService)

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, nil, resourceManager, false)
		require.NoError(t, err)
		assert.Empty(t, resourceManager.ListResources())
	})

	t.Run("disabled prompt", func(t *testing.T) {
		toolManager := NewMockToolManager()
		poolManager := pool.NewManager()
		promptManager := NewMockPromptManager()
		upstream := NewUpstream(poolManager)

		webrtcService := &configv1.WebrtcUpstreamService{}
		webrtcService.SetAddress("http://localhost:8080/signal")

		promptDef := &configv1.PromptDefinition{}
		promptDef.SetName("disabled-prompt")
		promptDef.SetDisable(true)
		webrtcService.SetPrompts([]*configv1.PromptDefinition{promptDef})

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-webrtc-disabled-prompt")
		serviceConfig.SetWebrtcService(webrtcService)

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, nil, false)
		require.NoError(t, err)
		assert.Empty(t, promptManager.ListPrompts())
	})
}
