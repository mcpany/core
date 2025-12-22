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
	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/client"
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

func NewMockToolManager(_ *bus.Provider) *MockToolManager {
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

func TestUpstream_Register_DisabledTool(t *testing.T) {
	toolManager := NewMockToolManager(nil)
	poolManager := pool.NewManager()
	var promptManager prompt.ManagerInterface
	var resourceManager resource.ManagerInterface
	upstream := NewUpstream(poolManager)

	toolDef := configv1.ToolDefinition_builder{
		Name:        proto.String("echo"),
		Description: proto.String("Echoes a message"),
		CallId:      proto.String("echo-call"),
		Disable:     proto.Bool(true),
	}.Build()

	websocketService := &configv1.WebsocketUpstreamService{}
	websocketService.SetAddress("ws://localhost:8080/echo")
	websocketService.SetTools([]*configv1.ToolDefinition{toolDef})
	calls := make(map[string]*configv1.WebsocketCallDefinition)
	calls["echo-call"] = configv1.WebsocketCallDefinition_builder{
		Id: proto.String("echo-call"),
	}.Build()
	websocketService.SetCalls(calls)

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-websocket-service")
	serviceConfig.SetWebsocketService(websocketService)

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
	require.NoError(t, err)

	tools := toolManager.ListTools()
	assert.Len(t, tools, 0)
}

func TestNewUpstream(t *testing.T) {
	poolManager := pool.NewManager()
	upstream := NewUpstream(poolManager)
	require.NotNil(t, upstream)
	assert.IsType(t, &Upstream{}, upstream)
}

func TestUpstream_Register_Mocked(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		toolManager := NewMockToolManager(nil)
		poolManager := pool.NewManager()
		var promptManager prompt.ManagerInterface
		var resourceManager resource.ManagerInterface

		upstream := NewUpstream(poolManager)

		toolDef := configv1.ToolDefinition_builder{
			Name:        proto.String("echo"),
			Description: proto.String("Echoes a message"),
			CallId:      proto.String("echo-call"),
		}.Build()

		websocketService := &configv1.WebsocketUpstreamService{}
		websocketService.SetAddress("ws://localhost:8080/echo")
		websocketService.SetTools([]*configv1.ToolDefinition{toolDef})
		calls := make(map[string]*configv1.WebsocketCallDefinition)
		calls["echo-call"] = configv1.WebsocketCallDefinition_builder{
			Id: proto.String("echo-call"),
		}.Build()
		websocketService.SetCalls(calls)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-websocket-service")
		serviceConfig.SetWebsocketService(websocketService)

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
		toolManager := NewMockToolManager(nil)
		poolManager := pool.NewManager()
		var promptManager prompt.ManagerInterface
		var resourceManager resource.ManagerInterface
		upstream := NewUpstream(poolManager)

		_, _, _, err := upstream.Register(context.Background(), nil, toolManager, promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "service config is nil")
	})

	t.Run("nil websocket service config", func(t *testing.T) {
		toolManager := NewMockToolManager(nil)
		poolManager := pool.NewManager()
		var promptManager prompt.ManagerInterface
		var resourceManager resource.ManagerInterface
		upstream := NewUpstream(poolManager)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-websocket-service")
		serviceConfig.SetWebsocketService(nil)

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Equal(t, "websocket service config is nil", err.Error())
	})

	t.Run("add tool error", func(t *testing.T) {
		toolManager := NewMockToolManager(nil)
		toolManager.lastErr = errors.New("failed to add tool")
		poolManager := pool.NewManager()
		var promptManager prompt.ManagerInterface
		var resourceManager resource.ManagerInterface
		upstream := NewUpstream(poolManager)

		toolDef := configv1.ToolDefinition_builder{
			Name:   proto.String("echo"),
			CallId: proto.String("echo-call"),
		}.Build()

		websocketService := &configv1.WebsocketUpstreamService{}
		websocketService.SetAddress("ws://localhost:8080/echo")
		websocketService.SetTools([]*configv1.ToolDefinition{toolDef})
		calls := make(map[string]*configv1.WebsocketCallDefinition)
		calls["echo-call"] = configv1.WebsocketCallDefinition_builder{
			Id: proto.String("echo-call"),
		}.Build()
		websocketService.SetCalls(calls)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-websocket-service")
		serviceConfig.SetWebsocketService(websocketService)

		_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)
		assert.Empty(t, discoveredTools)
	})

	t.Run("authenticator creation fails", func(t *testing.T) {
		toolManager := NewMockToolManager(nil)
		poolManager := pool.NewManager()
		var promptManager prompt.ManagerInterface
		var resourceManager resource.ManagerInterface
		upstream := NewUpstream(poolManager)

		toolDef := configv1.ToolDefinition_builder{
			Name:   proto.String("echo"),
			CallId: proto.String("echo-call"),
		}.Build()

		websocketService := &configv1.WebsocketUpstreamService{}
		websocketService.SetAddress("ws://localhost:8080/echo")
		websocketService.SetTools([]*configv1.ToolDefinition{toolDef})
		calls := make(map[string]*configv1.WebsocketCallDefinition)
		calls["echo-call"] = configv1.WebsocketCallDefinition_builder{
			Id: proto.String("echo-call"),
		}.Build()
		websocketService.SetCalls(calls)

		authConfig := (&configv1.UpstreamAuthentication_builder{
			ApiKey: &configv1.UpstreamAPIKeyAuth{},
		}).Build()

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("auth-fail-service")
		serviceConfig.SetWebsocketService(websocketService)
		serviceConfig.SetUpstreamAuthentication(authConfig)

		_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)
		assert.Len(t, discoveredTools, 1, "a tool should be discovered if auth config is incomplete")
	})

	t.Run("tool registration with fallback operation ID", func(t *testing.T) {
		toolManager := NewMockToolManager(nil)
		poolManager := pool.NewManager()
		var promptManager prompt.ManagerInterface
		var resourceManager resource.ManagerInterface
		upstream := NewUpstream(poolManager)

		// Fallback to description
		toolDef1 := configv1.ToolDefinition_builder{
			Description: proto.String("This is a test description"),
			CallId:      proto.String("call1"),
		}.Build()

		toolDef2 := configv1.ToolDefinition_builder{
			Description: proto.String(""),
			CallId:      proto.String("call2"),
		}.Build()

		websocketService := &configv1.WebsocketUpstreamService{}
		websocketService.SetAddress("ws://localhost:8080/echo")
		websocketService.SetTools([]*configv1.ToolDefinition{toolDef1, toolDef2})
		calls := make(map[string]*configv1.WebsocketCallDefinition)
		calls["call1"] = configv1.WebsocketCallDefinition_builder{
			Id: proto.String("call1"),
		}.Build()
		calls["call2"] = configv1.WebsocketCallDefinition_builder{
			Id: proto.String("call2"),
		}.Build()
		websocketService.SetCalls(calls)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-service-fallback")
		serviceConfig.SetWebsocketService(websocketService)

		serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)

		tools := toolManager.ListTools()
		assert.Len(t, tools, 2)

		sanitizedName := util.SanitizeOperationID("This is a test description")
		sanitizedName, _ = util.SanitizeToolName(sanitizedName)
		toolID1 := serviceID + "." + sanitizedName
		_, ok := toolManager.GetTool(toolID1)
		assert.True(t, ok, "Tool with sanitized description should be found, expected %s", toolID1)

		sanitizedName2, _ := util.SanitizeToolName("op1")
		toolID2 := serviceID + "." + sanitizedName2
		_, ok = toolManager.GetTool(toolID2)
		assert.True(t, ok, "tool should be registered with op index")
	})

	t.Run("correct input schema generation", func(t *testing.T) {
		toolManager := NewMockToolManager(nil)
		poolManager := pool.NewManager()
		upstream := NewUpstream(poolManager)

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

		toolDef := configv1.ToolDefinition_builder{
			Name:   proto.String("test-tool"),
			CallId: proto.String("test-call"),
		}.Build()

		websocketService := configv1.WebsocketUpstreamService_builder{
			Address: proto.String("ws://localhost:8080/test"),
			Tools:   []*configv1.ToolDefinition{toolDef},
			Calls: map[string]*configv1.WebsocketCallDefinition{
				"test-call": configv1.WebsocketCallDefinition_builder{
					Id:         proto.String("test-call"),
					Parameters: []*configv1.WebsocketParameterMapping{param1, param2},
				}.Build(),
			},
		}.Build()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name:             proto.String("test-websocket-service"),
			WebsocketService: websocketService,
		}.Build()

		serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, nil, nil, false)
		require.NoError(t, err)

		sanitizedToolName, _ := util.SanitizeToolName("test-tool")
		toolID := serviceID + "." + sanitizedToolName
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

func TestUpstream_Register_Integration(t *testing.T) {
	upgrader := websocket.Upgrader{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	poolManager := pool.NewManager()
	tm := tool.NewManager(nil)

	t.Run("successful registration", func(t *testing.T) {
		upstream := NewUpstream(poolManager)

		apiKeyAuth := &configv1.UpstreamAPIKeyAuth{}
		apiKeyAuth.SetHeaderName("X-API-Key")
		secret := &configv1.SecretValue{}
		secret.SetPlainText("test-key")
		apiKeyAuth.SetApiKey(secret)

		authConfig := &configv1.UpstreamAuthentication{}
		authConfig.SetApiKey(apiKeyAuth)

		tool1 := configv1.ToolDefinition_builder{
			Name:        proto.String("test-op"),
			Description: proto.String("A test operation"),
			CallId:      proto.String("call1"),
		}.Build()

		tool2 := configv1.ToolDefinition_builder{
			Description: proto.String("Another test operation"),
			CallId:      proto.String("call2"),
		}.Build()

		wsService := &configv1.WebsocketUpstreamService{}
		wsService.SetAddress(wsURL)
		wsService.SetTools([]*configv1.ToolDefinition{tool1, tool2})
		calls := make(map[string]*configv1.WebsocketCallDefinition)
		calls["call1"] = configv1.WebsocketCallDefinition_builder{
			Id: proto.String("call1"),
		}.Build()
		calls["call2"] = configv1.WebsocketCallDefinition_builder{
			Id: proto.String("call2"),
		}.Build()
		wsService.SetCalls(calls)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-service")
		serviceConfig.SetUpstreamAuthentication(authConfig)
		serviceConfig.SetWebsocketService(wsService)

		serviceID, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		require.NoError(t, err)
		expectedKey, _ := util.SanitizeServiceName("test-service")
		assert.Equal(t, expectedKey, serviceID)
		assert.Len(t, discoveredTools, 2)
		_, ok := pool.Get[*client.WebsocketClientWrapper](poolManager, serviceID)
		assert.True(t, ok)
	})

	t.Run("nil websocket service config", func(t *testing.T) {
		upstream := NewUpstream(poolManager)
		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("nil-config-service")

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "websocket service config is nil")
	})

	t.Run("invalid service name", func(t *testing.T) {
		upstream := NewUpstream(poolManager)
		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("")
		serviceConfig.SetWebsocketService(&configv1.WebsocketUpstreamService{})

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "id cannot be empty")
	})

	t.Run("authenticator creation fails", func(t *testing.T) {
		upstream := NewUpstream(poolManager)

		tool1 := configv1.ToolDefinition_builder{
			Name:   proto.String("test-op"),
			CallId: proto.String("test-call"),
		}.Build()

		wsService := &configv1.WebsocketUpstreamService{}
		wsService.SetAddress(wsURL)
		wsService.SetTools([]*configv1.ToolDefinition{tool1})
		calls := make(map[string]*configv1.WebsocketCallDefinition)
		calls["test-call"] = configv1.WebsocketCallDefinition_builder{
			Id: proto.String("test-call"),
		}.Build()
		wsService.SetCalls(calls)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("auth-fail-service")
		// Intentionally not setting auth method on UpstreamAuthentication, which is a valid scenario (no auth).
		serviceConfig.SetUpstreamAuthentication(&configv1.UpstreamAuthentication{})
		serviceConfig.SetWebsocketService(wsService)

		serviceID, tools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		require.NoError(t, err)
		expectedKey, _ := util.SanitizeServiceName("auth-fail-service")
		assert.Equal(t, expectedKey, serviceID)
		assert.Len(t, tools, 1, "expected one tool when authenticator is not configured")
	})

	t.Run("tool registration with fallback operation ID", func(t *testing.T) {
		tm := tool.NewManager(nil)
		upstream := NewUpstream(poolManager)

		tool1 := configv1.ToolDefinition_builder{
			Description: proto.String("A test operation"),
			CallId:      proto.String("call1"),
		}.Build()

		tool2 := configv1.ToolDefinition_builder{
			Description: proto.String("Another test operation"),
			CallId:      proto.String("call2"),
		}.Build()

		wsService := &configv1.WebsocketUpstreamService{}
		wsService.SetAddress(wsURL)
		wsService.SetTools([]*configv1.ToolDefinition{tool1, tool2})
		calls := make(map[string]*configv1.WebsocketCallDefinition)
		calls["call1"] = configv1.WebsocketCallDefinition_builder{
			Id: proto.String("call1"),
		}.Build()
		calls["call2"] = configv1.WebsocketCallDefinition_builder{
			Id: proto.String("call2"),
		}.Build()
		wsService.SetCalls(calls)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("fallback-op-id")
		serviceConfig.SetWebsocketService(wsService)

		serviceID, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		require.NoError(t, err)
		expectedKey, _ := util.SanitizeServiceName("fallback-op-id")
		assert.Equal(t, expectedKey, serviceID)
		assert.Len(t, discoveredTools, 2)
	})
}

func TestUpstream_Register_WithReload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
	}))
	defer server.Close()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	poolManager := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(poolManager)

	tool1 := configv1.ToolDefinition_builder{
		Name:   proto.String("test-op"),
		CallId: proto.String("test-call"),
	}.Build()

	wsService := &configv1.WebsocketUpstreamService{}
	wsService.SetAddress(wsURL)
	wsService.SetTools([]*configv1.ToolDefinition{tool1})
	calls := make(map[string]*configv1.WebsocketCallDefinition)
	calls["test-call"] = configv1.WebsocketCallDefinition_builder{
		Id: proto.String("test-call"),
	}.Build()
	wsService.SetCalls(calls)

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("reload-test")
	serviceConfig.SetWebsocketService(wsService)

	serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)
	sanitizedToolName, _ := util.SanitizeToolName("test-op")
	toolID := serviceID + "." + sanitizedToolName
	retrievedTool, ok := tm.GetTool(toolID)
	assert.True(t, ok)
	assert.NotNil(t, retrievedTool)

	_, _, _, err = upstream.Register(context.Background(), serviceConfig, tm, nil, nil, true)
	require.NoError(t, err)
	retrievedTool, ok = tm.GetTool(toolID)
	assert.True(t, ok)
	assert.NotNil(t, retrievedTool)
}

func TestUpstream_Register_DisabledItems(t *testing.T) {
	poolManager := pool.NewManager()
	tm := tool.NewManager(nil)
	pm := prompt.NewManager()
	rm := resource.NewManager()
	upstream := NewUpstream(poolManager)

	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	defer server.Close()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	enabledTool := configv1.ToolDefinition_builder{
		Name:   proto.String("enabled-tool"),
		CallId: proto.String("enabled-call"),
	}.Build()
	disabledTool := configv1.ToolDefinition_builder{
		Name:    proto.String("disabled-tool"),
		CallId:  proto.String("disabled-call"),
		Disable: proto.Bool(true),
	}.Build()

	enabledPrompt := &configv1.PromptDefinition{}
	enabledPrompt.SetName("enabled-prompt")
	disabledPrompt := &configv1.PromptDefinition{}
	disabledPrompt.SetName("disabled-prompt")
	disabledPrompt.SetDisable(true)

	wsService := &configv1.WebsocketUpstreamService{}
	wsService.SetAddress(wsURL)
	wsService.SetTools([]*configv1.ToolDefinition{enabledTool, disabledTool})
	wsService.SetCalls(map[string]*configv1.WebsocketCallDefinition{
		"enabled-call":  configv1.WebsocketCallDefinition_builder{Id: proto.String("enabled-call")}.Build(),
		"disabled-call": configv1.WebsocketCallDefinition_builder{Id: proto.String("disabled-call")}.Build(),
	})
	wsService.SetPrompts([]*configv1.PromptDefinition{enabledPrompt, disabledPrompt})

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("disabled-items-test")
	serviceConfig.SetWebsocketService(wsService)

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, pm, rm, false)
	require.NoError(t, err)

	assert.Len(t, tm.ListTools(), 1, "Only enabled tools should be registered")
	assert.Len(t, pm.ListPrompts(), 1, "Only enabled prompts should be registered")
}

func TestUpstream_Register_MissingCallDefinition(t *testing.T) {
	poolManager := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(poolManager)

	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	defer server.Close()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	toolWithMissingCall := configv1.ToolDefinition_builder{
		Name:   proto.String("tool-missing-call"),
		CallId: proto.String("missing-call"),
	}.Build()

	wsService := &configv1.WebsocketUpstreamService{}
	wsService.SetAddress(wsURL)
	wsService.SetTools([]*configv1.ToolDefinition{toolWithMissingCall})
	wsService.SetCalls(map[string]*configv1.WebsocketCallDefinition{})

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("missing-call-def-test")
	serviceConfig.SetWebsocketService(wsService)

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)
	assert.Empty(t, tm.ListTools(), "No tools should be registered if call definition is missing")
}
