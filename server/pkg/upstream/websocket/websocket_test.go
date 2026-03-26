// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// MockToolManager is a mock implementation of the ToolManagerInterface.
// MockToolManager is a mock implementation of the ToolManagerInterface.
// Summary: MockToolManager
	mu      sync.Mutex
	tools   map[string]tool.Tool
	lastErr error
}

// NewMockToolManager ...
// Summary: NewMockToolManager
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return &MockToolManager{
		tools: make(map[string]tool.Tool),
	}
}

// AddTool ...
// Summary: AddTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// GetTool ...
// Summary: GetTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tools[name]
	return t, ok
}

// IsServiceAllowed ...
// Summary: IsServiceAllowed
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return true
}

// ListTools ...
// Summary: ListTools
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.mu.Lock()
	defer m.mu.Unlock()
	tools := make([]tool.Tool, 0, len(m.tools))
	for _, t := range m.tools {
		tools = append(tools, t)
	}
	return tools
}

// ListMCPTools ...
// Summary: ListMCPTools
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}

// ClearToolsForService ...
// Summary: ClearToolsForService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, t := range m.tools {
		if t.Tool().GetServiceId() == serviceID {
			delete(m.tools, name)
		}
	}
}

// SetMCPServer ...
// Summary: SetMCPServer
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.

// SetProfiles ...
// Summary: SetProfiles
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.

// AddServiceInfo ...
// Summary: AddServiceInfo
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.

// GetServiceInfo ...
// Summary: GetServiceInfo
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, false
}

// ListServices ...
// Summary: ListServices
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}

// ExecuteTool ...
// Summary: ExecuteTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, errors.New("not implemented")
}

// AddMiddleware ...
// Summary: AddMiddleware
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
}

// ToolMatchesProfile ...
// Summary: ToolMatchesProfile
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return true
}

// TestUpstream_Register_DisabledTool ...
// Summary: TestUpstream_Register_DisabledTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

	websocketService := configv1.WebsocketUpstreamService_builder{
		Address: proto.String("ws://127.0.0.1:8080/echo"),
		Tools:   []*configv1.ToolDefinition{toolDef},
		Calls: map[string]*configv1.WebsocketCallDefinition{
			"echo-call": configv1.WebsocketCallDefinition_builder{
				Id: proto.String("echo-call"),
			}.Build(),
		},
	}.Build()

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name:             proto.String("test-websocket-service"),
		WebsocketService: websocketService,
	}.Build()

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
	require.NoError(t, err)

	tools := toolManager.ListTools()
	assert.Len(t, tools, 0)
}

// TestNewUpstream ...
// Summary: TestNewUpstream
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	poolManager := pool.NewManager()
	upstream := NewUpstream(poolManager)
	require.NotNil(t, upstream)
	assert.IsType(t, &Upstream{}, upstream)
}

// TestUpstream_Register_Mocked ...
// Summary: TestUpstream_Register_Mocked
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

		websocketService := configv1.WebsocketUpstreamService_builder{
			Address: proto.String("ws://127.0.0.1:8080/echo"),
			Tools:   []*configv1.ToolDefinition{toolDef},
			Calls: map[string]*configv1.WebsocketCallDefinition{
				"echo-call": configv1.WebsocketCallDefinition_builder{
					Id: proto.String("echo-call"),
				}.Build(),
			},
		}.Build()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name:             proto.String("test-websocket-service"),
			WebsocketService: websocketService,
		}.Build()

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

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test-websocket-service"),
		}.Build()

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

		websocketService := configv1.WebsocketUpstreamService_builder{
			Address: proto.String("ws://127.0.0.1:8080/echo"),
			Tools:   []*configv1.ToolDefinition{toolDef},
			Calls: map[string]*configv1.WebsocketCallDefinition{
				"echo-call": configv1.WebsocketCallDefinition_builder{
					Id: proto.String("echo-call"),
				}.Build(),
			},
		}.Build()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name:             proto.String("test-websocket-service"),
			WebsocketService: websocketService,
		}.Build()

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

		websocketService := configv1.WebsocketUpstreamService_builder{
			Address: proto.String("ws://127.0.0.1:8080/echo"),
			Tools:   []*configv1.ToolDefinition{toolDef},
			Calls: map[string]*configv1.WebsocketCallDefinition{
				"echo-call": configv1.WebsocketCallDefinition_builder{
					Id: proto.String("echo-call"),
				}.Build(),
			},
		}.Build()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name:             proto.String("auth-fail-service"),
			WebsocketService: websocketService,
			UpstreamAuth: configv1.Authentication_builder{
				ApiKey: configv1.APIKeyAuth_builder{}.Build(),
			}.Build(),
		}.Build()

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

		websocketService := configv1.WebsocketUpstreamService_builder{
			Address: proto.String("ws://127.0.0.1:8080/echo"),
			Tools:   []*configv1.ToolDefinition{toolDef1, toolDef2},
			Calls: map[string]*configv1.WebsocketCallDefinition{
				"call1": configv1.WebsocketCallDefinition_builder{
					Id: proto.String("call1"),
				}.Build(),
				"call2": configv1.WebsocketCallDefinition_builder{
					Id: proto.String("call2"),
				}.Build(),
			},
		}.Build()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name:             proto.String("test-service-fallback"),
			WebsocketService: websocketService,
		}.Build()

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
				Name:       proto.String("param1"),
				IsRequired: proto.Bool(true),
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
			Address: proto.String("ws://127.0.0.1:8080/test"),
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

		requiredVal, ok := inputSchema.GetFields()["required"]
		require.True(t, ok, "required field should be present")
		requiredList := requiredVal.GetListValue().GetValues()
		assert.Len(t, requiredList, 1)
		assert.Equal(t, "param1", requiredList[0].GetStringValue())
	})
}

// TestUpstream_Register_Integration ...
// Summary: TestUpstream_Register_Integration
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

		apiKeyAuth := configv1.APIKeyAuth_builder{
			ParamName: proto.String("X-API-Key"),
			Value: configv1.SecretValue_builder{
				PlainText: proto.String("test-key"),
			}.Build(),
		}.Build()

		authConfig := configv1.Authentication_builder{
			ApiKey: apiKeyAuth,
		}.Build()

		tool1 := configv1.ToolDefinition_builder{
			Name:        proto.String("test-op"),
			Description: proto.String("A test operation"),
			CallId:      proto.String("call1"),
		}.Build()

		tool2 := configv1.ToolDefinition_builder{
			Description: proto.String("Another test operation"),
			CallId:      proto.String("call2"),
		}.Build()

		wsService := configv1.WebsocketUpstreamService_builder{
			Address: proto.String(wsURL),
			Tools:   []*configv1.ToolDefinition{tool1, tool2},
			Calls: map[string]*configv1.WebsocketCallDefinition{
				"call1": configv1.WebsocketCallDefinition_builder{
					Id: proto.String("call1"),
				}.Build(),
				"call2": configv1.WebsocketCallDefinition_builder{
					Id: proto.String("call2"),
				}.Build(),
			},
		}.Build()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name:             proto.String("test-service"),
			WebsocketService: wsService,
			UpstreamAuth:     authConfig,
		}.Build()

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
		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("nil-config-service"),
		}.Build()

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "websocket service config is nil")
	})

	t.Run("invalid service name", func(t *testing.T) {
		upstream := NewUpstream(poolManager)
		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name:             proto.String(""),
			WebsocketService: configv1.WebsocketUpstreamService_builder{}.Build(),
		}.Build()

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

		wsService := configv1.WebsocketUpstreamService_builder{
			Address: proto.String(wsURL),
			Tools:   []*configv1.ToolDefinition{tool1},
			Calls: map[string]*configv1.WebsocketCallDefinition{
				"test-call": configv1.WebsocketCallDefinition_builder{
					Id: proto.String("test-call"),
				}.Build(),
			},
		}.Build()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name:             proto.String("auth-fail-service"),
			WebsocketService: wsService,
			UpstreamAuth:     configv1.Authentication_builder{}.Build(),
		}.Build()

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

		wsService := configv1.WebsocketUpstreamService_builder{
			Address: proto.String(wsURL),
			Tools:   []*configv1.ToolDefinition{tool1, tool2},
			Calls: map[string]*configv1.WebsocketCallDefinition{
				"call1": configv1.WebsocketCallDefinition_builder{
					Id: proto.String("call1"),
				}.Build(),
				"call2": configv1.WebsocketCallDefinition_builder{
					Id: proto.String("call2"),
				}.Build(),
			},
		}.Build()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name:             proto.String("fallback-op-id"),
			WebsocketService: wsService,
		}.Build()

		serviceID, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		require.NoError(t, err)
		expectedKey, _ := util.SanitizeServiceName("fallback-op-id")
		assert.Equal(t, expectedKey, serviceID)
		assert.Len(t, discoveredTools, 2)
	})
}

// TestUpstream_Register_WithReload ...
// Summary: TestUpstream_Register_WithReload
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

	wsService := configv1.WebsocketUpstreamService_builder{
		Address: proto.String(wsURL),
		Tools:   []*configv1.ToolDefinition{tool1},
		Calls: map[string]*configv1.WebsocketCallDefinition{
			"test-call": configv1.WebsocketCallDefinition_builder{
				Id: proto.String("test-call"),
			}.Build(),
		},
	}.Build()

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name:             proto.String("reload-test"),
		WebsocketService: wsService,
	}.Build()

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

// TestUpstream_Register_DisabledItems ...
// Summary: TestUpstream_Register_DisabledItems
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

	enabledPrompt := configv1.PromptDefinition_builder{
		Name: proto.String("enabled-prompt"),
	}.Build()
	disabledPrompt := configv1.PromptDefinition_builder{
		Name:    proto.String("disabled-prompt"),
		Disable: proto.Bool(true),
	}.Build()

	wsService := configv1.WebsocketUpstreamService_builder{
		Address: proto.String(wsURL),
		Tools:   []*configv1.ToolDefinition{enabledTool, disabledTool},
		Calls: map[string]*configv1.WebsocketCallDefinition{
			"enabled-call":  configv1.WebsocketCallDefinition_builder{Id: proto.String("enabled-call")}.Build(),
			"disabled-call": configv1.WebsocketCallDefinition_builder{Id: proto.String("disabled-call")}.Build(),
		},
		Prompts: []*configv1.PromptDefinition{enabledPrompt, disabledPrompt},
	}.Build()

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name:             proto.String("disabled-items-test"),
		WebsocketService: wsService,
	}.Build()

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, pm, rm, false)
	require.NoError(t, err)

	assert.Len(t, tm.ListTools(), 1, "Only enabled tools should be registered")
	assert.Len(t, pm.ListPrompts(), 1, "Only enabled prompts should be registered")
}

// TestUpstream_Register_MissingCallDefinition ...
// Summary: TestUpstream_Register_MissingCallDefinition
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

	wsService := configv1.WebsocketUpstreamService_builder{
		Address: proto.String(wsURL),
		Tools:   []*configv1.ToolDefinition{toolWithMissingCall},
	}.Build()

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name:             proto.String("missing-call-def-test"),
		WebsocketService: wsService,
	}.Build()

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)
	assert.Empty(t, tm.ListTools(), "No tools should be registered if call definition is missing")
}

// TestUpstream_createAndRegisterWebsocketTools_DynamicResourceMissingTool ...
// Summary: TestUpstream_createAndRegisterWebsocketTools_DynamicResourceMissingTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	toolManager := tool.NewManager(nil)
	resourceManager := resource.NewManager()
	poolManager := pool.NewManager()
	upstream := NewUpstream(poolManager)

	dynamicResource := configv1.ResourceDefinition_builder{
		Name: proto.String("test-resource"),
		Dynamic: configv1.DynamicResource_builder{
			WebsocketCall: configv1.WebsocketCallDefinition_builder{
				Id: proto.String("missing-tool"),
			}.Build(),
		}.Build(),
	}.Build()

	websocketService := configv1.WebsocketUpstreamService_builder{
		Address:   proto.String("ws://127.0.0.1:8080/test"),
		Resources: []*configv1.ResourceDefinition{dynamicResource},
	}.Build()

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name:             proto.String("test-websocket-service"),
		WebsocketService: websocketService,
	}.Build()

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, nil, resourceManager, false)
	require.NoError(t, err)
	assert.Empty(t, resourceManager.ListResources(), "No resources should be registered if tool is missing")
}

// GetAllowedServiceIDs ...
// Summary: GetAllowedServiceIDs
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, true
}

// GetToolCountForService ...
// Summary: GetToolCountForService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return 0
}
