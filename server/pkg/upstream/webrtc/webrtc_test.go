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

// MockPromptManager is a mock implementation of the PromptManagerInterface.
// MockPromptManager is a mock implementation of the PromptManagerInterface.
// Summary: MockPromptManager
	mu      sync.Mutex
	prompts map[string]prompt.Prompt
}

// NewMockPromptManager ...
// Summary: NewMockPromptManager
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return &MockPromptManager{
		prompts: make(map[string]prompt.Prompt),
	}
}

// AddPrompt ...
// Summary: AddPrompt
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
	m.prompts[p.Prompt().Name] = p
}

// UpdatePrompt ...
// Summary: UpdatePrompt
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
	m.prompts[p.Prompt().Name] = p
}

// GetPrompt ...
// Summary: GetPrompt
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
	p, ok := m.prompts[name]
	return p, ok
}

// ListPrompts ...
// Summary: ListPrompts
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
	prompts := make([]prompt.Prompt, 0, len(m.prompts))
	for _, p := range m.prompts {
		prompts = append(prompts, p)
	}
	return prompts
}

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

// ClearPromptsForService ...
// Summary: ClearPromptsForService
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
	for name, p := range m.prompts {
		if p.Service() == serviceID {
			delete(m.prompts, name)
		}
	}
}

// MockResourceManager is a mock implementation of the ResourceManagerInterface.
// MockResourceManager is a mock implementation of the ResourceManagerInterface.
// Summary: MockResourceManager
	mu        sync.Mutex
	resources map[string]resource.Resource
}

// NewMockResourceManager ...
// Summary: NewMockResourceManager
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return &MockResourceManager{
		resources: make(map[string]resource.Resource),
	}
}

// AddResource ...
// Summary: AddResource
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
	m.resources[r.Resource().Name] = r
}

// GetResource ...
// Summary: GetResource
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
	r, ok := m.resources[name]
	return r, ok
}

// RemoveResource ...
// Summary: RemoveResource
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.

// ListResources ...
// Summary: ListResources
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
	resources := make([]resource.Resource, 0, len(m.resources))
	for _, r := range m.resources {
		resources = append(resources, r)
	}
	return resources
}

// OnListChanged ...
// Summary: OnListChanged
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.

// ClearResourcesForService ...
// Summary: ClearResourcesForService
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
	for name, r := range m.resources {
		if r.Service() == serviceID {
			delete(m.resources, name)
		}
	}
}

// TestUpstream ...
// Summary: TestUpstream
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

// TestWebrtcUpstream_Shutdown ...
// Summary: TestWebrtcUpstream_Shutdown
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	u := NewUpstream(nil)
	assert.NotNil(t, u)

	err := u.Shutdown(context.Background())
	assert.NoError(t, err)
}

// TestUpstream_Register ...
// Summary: TestUpstream_Register
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestUpstream_Register_ToolNameGeneration ...
// Summary: TestUpstream_Register_ToolNameGeneration
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestUpstream_Register_CornerCases ...
// Summary: TestUpstream_Register_CornerCases
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
