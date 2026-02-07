// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package webrtc

import (
	"context"
	"errors"
	"sync"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	mcp_sdk "github.com/modelcontextprotocol/go-sdk/mcp"
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

func (m *MockToolManager) IsServiceAllowed(serviceID, profileID string) bool {
	return true
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

func (m *MockToolManager) ListMCPTools() []*mcp_sdk.Tool {
	return nil
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
func (m *MockToolManager) AddMiddleware(_ tool.ExecutionMiddleware) {}
func (m *MockToolManager) ToolMatchesProfile(tool tool.Tool, profileID string) bool {
	return true
}
func (m *MockToolManager) GetAllowedServiceIDs(_ string) (map[string]bool, bool) {
	return nil, true
}

func (m *MockToolManager) OnListChanged(f func()) {}

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

func (m *MockPromptManager) SetMCPServer(_ prompt.MCPServerProvider) {}

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

func (m *MockResourceManager) RemoveResource(_ string) {}

func (m *MockResourceManager) ListResources() []resource.Resource {
	m.mu.Lock()
	defer m.mu.Unlock()
	resources := make([]resource.Resource, 0, len(m.resources))
	for _, r := range m.resources {
		resources = append(resources, r)
	}
	return resources
}

func (m *MockResourceManager) OnListChanged(_ func()) {}

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

func TestWebrtcUpstream_Shutdown(t *testing.T) {
	u := NewUpstream(nil)
	assert.NotNil(t, u)

	err := u.Shutdown(context.Background())
	assert.NoError(t, err)
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

		webrtcService := configv1.WebrtcUpstreamService_builder{
			Address: proto.String("http://127.0.0.1:8080/signal"),
			Tools:   []*configv1.ToolDefinition{toolDef},
			Calls: map[string]*configv1.WebrtcCallDefinition{
				"echo-call": configv1.WebrtcCallDefinition_builder{
					Id: proto.String("echo-call"),
				}.Build(),
			},
		}.Build()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name:          proto.String("test-webrtc-service"),
			WebrtcService: webrtcService,
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

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test-webrtc-service"),
		}.Build()

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

		webrtcService := configv1.WebrtcUpstreamService_builder{
			Address: proto.String("http://127.0.0.1:8080/signal"),
			Tools:   []*configv1.ToolDefinition{toolDef},
			Calls: map[string]*configv1.WebrtcCallDefinition{
				"echo-call": configv1.WebrtcCallDefinition_builder{
					Id: proto.String("echo-call"),
				}.Build(),
			},
		}.Build()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name:          proto.String("test-webrtc-service"),
			WebrtcService: webrtcService,
		}.Build()

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

		authConfig := configv1.Authentication_builder{
			ApiKey: configv1.APIKeyAuth_builder{
				ParamName: proto.String(""), // Invalid header name
			}.Build(),
		}.Build()

		webrtcService := configv1.WebrtcUpstreamService_builder{
			Address: proto.String("http://127.0.0.1:8080/signal"),
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name:   proto.String("echo"),
					CallId: proto.String("echo-call"),
				}.Build(),
			},
			Calls: map[string]*configv1.WebrtcCallDefinition{
				"echo-call": configv1.WebrtcCallDefinition_builder{
					Id: proto.String("echo-call"),
				}.Build(),
			},
		}.Build()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name:          proto.String("test-webrtc-service"),
			UpstreamAuth:  authConfig,
			WebrtcService: webrtcService,
		}.Build()

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

		webrtcService := configv1.WebrtcUpstreamService_builder{
			Address: proto.String("http://127.0.0.1:8080/signal"),
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name:   proto.String("echo"),
					CallId: proto.String("non-existent-call-id"),
				}.Build(),
			},
		}.Build()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name:          proto.String("test-webrtc-service"),
			WebrtcService: webrtcService,
		}.Build()

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

		webrtcService := configv1.WebrtcUpstreamService_builder{
			Address: proto.String("http://127.0.0.1:8080/signal"),
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name:   proto.String("get-weather"),
					CallId: proto.String("get-weather-call"),
				}.Build(),
			},
			Calls: map[string]*configv1.WebrtcCallDefinition{
				"get-weather-call": configv1.WebrtcCallDefinition_builder{
					Id: proto.String("get-weather-call"),
				}.Build(),
			},
			Prompts: []*configv1.PromptDefinition{
				configv1.PromptDefinition_builder{
					Name: proto.String("weather-prompt"),
					Messages: []*configv1.PromptMessage{
						configv1.PromptMessage_builder{
							Text: configv1.TextContent_builder{
								Text: proto.String("What is the weather in {{.location}}?"),
							}.Build(),
						}.Build(),
					},
				}.Build(),
			},
			Resources: []*configv1.ResourceDefinition{
				configv1.ResourceDefinition_builder{
					Name: proto.String("weather-resource"),
					Dynamic: configv1.DynamicResource_builder{
						WebrtcCall: configv1.WebrtcCallDefinition_builder{
							Id: proto.String("get-weather-call"),
						}.Build(),
					}.Build(),
				}.Build(),
			},
		}.Build()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name:          proto.String("test-webrtc-service-with-prompts-and-resources"),
			WebrtcService: webrtcService,
		}.Build()

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

		webrtcService := configv1.WebrtcUpstreamService_builder{
			Address: proto.String("http://127.0.0.1:8080/signal"),
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name:   proto.String("get-weather"),
					CallId: proto.String("get-weather-call"),
				}.Build(),
			},
			Calls: map[string]*configv1.WebrtcCallDefinition{
				"get-weather-call": configv1.WebrtcCallDefinition_builder{
					Id: proto.String("get-weather-call"),
				}.Build(),
			},
			Resources: []*configv1.ResourceDefinition{
				configv1.ResourceDefinition_builder{
					Name: proto.String("weather-resource"),
					Dynamic: configv1.DynamicResource_builder{
						WebrtcCall: configv1.WebrtcCallDefinition_builder{
							Id: proto.String("get-weather-call"),
						}.Build(),
					}.Build(),
				}.Build(),
			},
		}.Build()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name:          proto.String("test-webrtc-service-with-sanitizer-failure"),
			WebrtcService: webrtcService,
		}.Build()

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

	webrtcService := configv1.WebrtcUpstreamService_builder{
		Address: proto.String("http://127.0.0.1:8080/signal"),
		Tools:   []*configv1.ToolDefinition{toolDef},
		Calls: map[string]*configv1.WebrtcCallDefinition{
			"test-call": configv1.WebrtcCallDefinition_builder{
				Id: proto.String("test-call"),
			}.Build(),
		},
	}.Build()

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name:          proto.String("test-webrtc-service-tool-name-generation"),
		WebrtcService: webrtcService,
	}.Build()

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

		webrtcService := configv1.WebrtcUpstreamService_builder{
			Address: proto.String("http://127.0.0.1:8080/signal"),
			Tools:   []*configv1.ToolDefinition{toolDef},
		}.Build()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name:          proto.String("test-webrtc-disabled"),
			WebrtcService: webrtcService,
		}.Build()

		_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, nil, nil, false)
		require.NoError(t, err)
		assert.Empty(t, discoveredTools)
	})

	t.Run("empty name fallback", func(t *testing.T) {
		toolManager := NewMockToolManager()
		poolManager := pool.NewManager()
		upstream := NewUpstream(poolManager)

		toolDef := configv1.ToolDefinition_builder{
			Name:        proto.String(""),
			Description: proto.String(""),
			CallId:      proto.String("call-id"),
		}.Build()

		webrtcService := configv1.WebrtcUpstreamService_builder{
			Address: proto.String("http://127.0.0.1:8080/signal"),
			Tools:   []*configv1.ToolDefinition{toolDef},
			Calls: map[string]*configv1.WebrtcCallDefinition{
				"call-id": configv1.WebrtcCallDefinition_builder{
					Id: proto.String("call-id"),
				}.Build(),
			},
		}.Build()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name:          proto.String("test-webrtc-empty-name"),
			WebrtcService: webrtcService,
		}.Build()

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

		webrtcService := configv1.WebrtcUpstreamService_builder{
			Address: proto.String("http://127.0.0.1:8080/signal"),
			Resources: []*configv1.ResourceDefinition{
				configv1.ResourceDefinition_builder{
					Name:    proto.String("disabled-resource"),
					Disable: proto.Bool(true),
				}.Build(),
			},
		}.Build()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name:          proto.String("test-webrtc-disabled-resource"),
			WebrtcService: webrtcService,
		}.Build()

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, nil, resourceManager, false)
		require.NoError(t, err)
		assert.Empty(t, resourceManager.ListResources())
	})

	t.Run("dynamic resource missing call", func(t *testing.T) {
		toolManager := NewMockToolManager()
		poolManager := pool.NewManager()
		resourceManager := NewMockResourceManager()
		upstream := NewUpstream(poolManager)

		webrtcService := configv1.WebrtcUpstreamService_builder{
			Address: proto.String("http://127.0.0.1:8080/signal"),
			Resources: []*configv1.ResourceDefinition{
				configv1.ResourceDefinition_builder{
					Name:    proto.String("resource-missing-call"),
					Dynamic: configv1.DynamicResource_builder{}.Build(),
				}.Build(),
			},
		}.Build()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name:          proto.String("test-webrtc-missing-call"),
			WebrtcService: webrtcService,
		}.Build()

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, nil, resourceManager, false)
		require.NoError(t, err)
		assert.Empty(t, resourceManager.ListResources())
	})

	t.Run("dynamic resource call id not found", func(t *testing.T) {
		toolManager := NewMockToolManager()
		poolManager := pool.NewManager()
		resourceManager := NewMockResourceManager()
		upstream := NewUpstream(poolManager)

		webrtcService := configv1.WebrtcUpstreamService_builder{
			Address: proto.String("http://127.0.0.1:8080/signal"),
			Resources: []*configv1.ResourceDefinition{
				configv1.ResourceDefinition_builder{
					Name: proto.String("resource-call-not-found"),
					Dynamic: configv1.DynamicResource_builder{
						WebrtcCall: configv1.WebrtcCallDefinition_builder{
							Id: proto.String("unknown-call-id"),
						}.Build(),
					}.Build(),
				}.Build(),
			},
		}.Build()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name:          proto.String("test-webrtc-call-not-found"),
			WebrtcService: webrtcService,
		}.Build()

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

		webrtcService := configv1.WebrtcUpstreamService_builder{
			Address: proto.String("http://127.0.0.1:8080/signal"),
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name:   proto.String("tool1"),
					CallId: proto.String("call1"),
				}.Build(),
			},
			Calls: map[string]*configv1.WebrtcCallDefinition{
				"call1": configv1.WebrtcCallDefinition_builder{
					Id: proto.String("call1"),
				}.Build(),
			},
			Resources: []*configv1.ResourceDefinition{
				configv1.ResourceDefinition_builder{
					Name: proto.String("resource1"),
					Dynamic: configv1.DynamicResource_builder{
						WebrtcCall: configv1.WebrtcCallDefinition_builder{
							Id: proto.String("call1"),
						}.Build(),
					}.Build(),
				}.Build(),
			},
		}.Build()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name:          proto.String("test-webrtc-tool-not-found"),
			WebrtcService: webrtcService,
		}.Build()

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, nil, resourceManager, false)
		require.NoError(t, err)
		assert.Empty(t, resourceManager.ListResources())
	})

	t.Run("disabled prompt", func(t *testing.T) {
		toolManager := NewMockToolManager()
		poolManager := pool.NewManager()
		promptManager := NewMockPromptManager()
		upstream := NewUpstream(poolManager)

		webrtcService := configv1.WebrtcUpstreamService_builder{
			Address: proto.String("http://127.0.0.1:8080/signal"),
			Prompts: []*configv1.PromptDefinition{
				configv1.PromptDefinition_builder{
					Name:    proto.String("disabled-prompt"),
					Disable: proto.Bool(true),
				}.Build(),
			},
		}.Build()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name:          proto.String("test-webrtc-disabled-prompt"),
			WebrtcService: webrtcService,
		}.Build()

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, nil, false)
		require.NoError(t, err)
		assert.Empty(t, promptManager.ListPrompts())
	})

	t.Run("correct input schema generation", func(t *testing.T) {
		toolManager := NewMockToolManager()
		poolManager := pool.NewManager()
		upstream := NewUpstream(poolManager)

		param1 := configv1.WebrtcParameterMapping_builder{
			Schema: configv1.ParameterSchema_builder{
				Name:       proto.String("param1"),
				IsRequired: proto.Bool(true),
			}.Build(),
		}.Build()
		param2 := configv1.WebrtcParameterMapping_builder{
			Schema: configv1.ParameterSchema_builder{
				Name: proto.String("param2"),
			}.Build(),
		}.Build()

		toolDef := configv1.ToolDefinition_builder{
			Name:   proto.String("test-tool"),
			CallId: proto.String("test-call"),
		}.Build()

		webrtcService := configv1.WebrtcUpstreamService_builder{
			Address: proto.String("http://localhost:8080/signal"),
			Tools:   []*configv1.ToolDefinition{toolDef},
			Calls: map[string]*configv1.WebrtcCallDefinition{
				"test-call": configv1.WebrtcCallDefinition_builder{
					Id:         proto.String("test-call"),
					Parameters: []*configv1.WebrtcParameterMapping{param1, param2},
				}.Build(),
			},
		}.Build()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name:          proto.String("test-webrtc-service"),
			WebrtcService: webrtcService,
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
