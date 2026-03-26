// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package serviceregistry

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	mcp_routerv1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// Mocks

type mockUpstream struct {
	upstream.Upstream
	registerFunc func(serviceName string) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error)
	shutdownFunc func() error
}

// Shutdown ...
// Summary: Shutdown
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.shutdownFunc != nil {
		return m.shutdownFunc()
	}
	return nil
}

// Register ...
// Summary: Register
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.registerFunc != nil {
		return m.registerFunc(serviceConfig.GetName())
	}
	return "mock-service-key", nil, nil, nil
}

type mockFactory struct {
	factory.Factory
	newUpstreamFunc func() (upstream.Upstream, error)
}

// NewUpstream ...
// Summary: NewUpstream
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.newUpstreamFunc != nil {
		return m.newUpstreamFunc()
	}
	return &mockUpstream{}, nil
}

type mockToolManager struct {
	tool.ManagerInterface
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
	return nil, nil
}
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

// TestNew ...
// Summary: TestNew
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	pm := pool.NewManager()
	f := factory.NewUpstreamServiceFactory(pm, nil)
	tm := &mockToolManager{}
	prm := prompt.NewManager()
	rm := resource.NewManager()
	am := auth.NewManager()

	registry := New(f, tm, prm, rm, am)
	assert.NotNil(t, registry)
	assert.NotNil(t, registry.serviceConfigs)
	assert.Equal(t, f, registry.factory)
	assert.Equal(t, tm, registry.toolManager)
	assert.Equal(t, prm, registry.promptManager)
	assert.Equal(t, rm, registry.resourceManager)
	assert.Equal(t, am, registry.authManager)
}

// TestServiceRegistry_RegisterAndGetService ...
// Summary: TestServiceRegistry_RegisterAndGetService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return &mockUpstream{
				registerFunc: func(serviceName string) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
					serviceID, err := util.SanitizeServiceName(serviceName)
					require.NoError(t, err)
					return serviceID, nil, nil, nil
				},
			}, nil
		},
	}
	tm := &mockToolManager{}
	prm := prompt.NewManager()
	rm := resource.NewManager()
	am := auth.NewManager()
	registry := New(f, tm, prm, rm, am)

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String("http://127.0.0.1"),
		}.Build(),
		Authentication: configv1.Authentication_builder{
			ApiKey: configv1.APIKeyAuth_builder{
				VerificationValue: proto.String("test-key"),
			}.Build(),
		}.Build(),
	}.Build()

	// Successful registration
	serviceID, tools, resources, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err)
	expectedServiceID, err := util.SanitizeServiceName("test-service")
	require.NoError(t, err)
	assert.Equal(t, expectedServiceID, serviceID)
	assert.Nil(t, tools)
	assert.Nil(t, resources)

	// Get the config
	retrievedConfig, ok := registry.GetServiceConfig(serviceID)
	require.True(t, ok)

	// Secrets should be stripped from retrieved config
	expectedConfig := proto.Clone(serviceConfig).(*configv1.UpstreamServiceConfig)
	util.StripSecretsFromService(expectedConfig)
	// Expect ToolCount to be populated (0 in this case)
	expectedConfig.SetToolCount(0)
	assert.Equal(t, expectedConfig, retrievedConfig)
	assert.Empty(t, retrievedConfig.GetAuthentication().GetApiKey().GetVerificationValue())

	// Check authenticator
	_, ok = am.GetAuthenticator(serviceID)
	assert.True(t, ok, "Authenticator should have been added")

	// Get non-existent config
	_, ok = registry.GetServiceConfig("non-existent")
	assert.False(t, ok)

	t.Run("with OAuth2 authenticator", func(t *testing.T) {
		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("oauth2-service"),
			HttpService: configv1.HttpUpstreamService_builder{
				Address: proto.String("http://127.0.0.1"),
			}.Build(),
			Authentication: configv1.Authentication_builder{
				Oauth2: configv1.OAuth2Auth_builder{
					IssuerUrl: proto.String("https://accounts.google.com"),
					Audience:  proto.String("test-audience"),
				}.Build(),
			}.Build(),
		}.Build()
		serviceID, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
		require.NoError(t, err)

		_, ok := am.GetAuthenticator(serviceID)
		assert.True(t, ok, "OAuth2 authenticator should have been added")
	})
}

// TestServiceRegistry_RegisterService_FactoryError ...
// Summary: TestServiceRegistry_RegisterService_FactoryError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	factoryErr := errors.New("factory error")
	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return nil, factoryErr
		},
	}
	tm := &mockToolManager{}
	registry := New(f, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
	}.Build()
	_, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), factoryErr.Error())
}

// TestServiceRegistry_RegisterService_UpstreamError ...
// Summary: TestServiceRegistry_RegisterService_UpstreamError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	upstreamErr := errors.New("upstream error")
	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return &mockUpstream{
				registerFunc: func(_ string) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
					return "", nil, nil, upstreamErr
				},
			}, nil
		},
	}
	tm := &mockToolManager{}
	registry := New(f, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-service")
	_, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), upstreamErr.Error())
}

// TestServiceRegistry_RegisterService_DuplicateName ...
// Summary: TestServiceRegistry_RegisterService_DuplicateName
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return &mockUpstream{
				registerFunc: func(serviceName string) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
					serviceID, err := util.SanitizeServiceName(serviceName)
					require.NoError(t, err)
					return serviceID, nil, nil, nil
				},
			}, nil
		},
	}
	tm := &mockToolManager{}
	registry := New(f, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	serviceConfig1 := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
	}.Build()
	_, _, _, err := registry.RegisterService(context.Background(), serviceConfig1)
	require.NoError(t, err, "First registration should succeed")

	// Attempt to register another service with the same name
	serviceConfig2 := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
	}.Build()
	_, _, _, err = registry.RegisterService(context.Background(), serviceConfig2)
	require.Error(t, err, "Second registration with the same name should fail")
	assert.ErrorIs(t, err, ErrServiceAlreadyRegistered)
}

// TestServiceRegistry_UnregisterService ...
// Summary: TestServiceRegistry_UnregisterService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return &mockUpstream{
				registerFunc: func(serviceName string) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
					serviceID, err := util.SanitizeServiceName(serviceName)
					require.NoError(t, err)
					return serviceID, nil, nil, nil
				},
			}, nil
		},
	}
	tm := &mockToolManager{}
	prm := prompt.NewManager()
	rm := resource.NewManager()
	am := auth.NewManager()
	registry := New(f, tm, prm, rm, am)

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
	}.Build()

	// Register the service
	serviceID, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err)
	expectedServiceID, err := util.SanitizeServiceName("test-service")
	require.NoError(t, err)
	assert.Equal(t, expectedServiceID, serviceID)

	// Verify it's registered
	_, ok := registry.GetServiceConfig(serviceID)
	require.True(t, ok, "Service should be registered before unregistering")

	// Unregister the service
	err = registry.UnregisterService(context.Background(), serviceID)
	require.NoError(t, err)

	// Verify it's no longer registered
	_, ok = registry.GetServiceConfig(serviceID)
	assert.False(t, ok, "Service should not be registered after unregistering")

	// Try to unregister a non-existent service
	err = registry.UnregisterService(context.Background(), "non-existent-service")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `service "non-existent-service" (id: non-existent-service) not found`)
}

// TestServiceRegistry_Close ...
// Summary: TestServiceRegistry_Close
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	mockUpstream := &mockUpstream{
		registerFunc: func(serviceName string) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
			serviceID, err := util.SanitizeServiceName(serviceName)
			if err != nil {
				return "", nil, nil, err
			}
			return serviceID, nil, nil, nil
		},
	}
	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return mockUpstream, nil
		},
	}
	tm := &mockToolManager{}
	registry := New(f, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	// Register a service
	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
	}.Build()
	_, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err)

	// Successful Close
	err = registry.Close(context.Background())
	require.NoError(t, err)

	// Close with error
	mockUpstream.shutdownFunc = func() error {
		return errors.New("shutdown error")
	}
	// Register another service to test error
	serviceConfig2 := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("shutdown-test-service"),
	}.Build()
	_, _, _, err = registry.RegisterService(context.Background(), serviceConfig2)
	require.NoError(t, err)

	err = registry.Close(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "shutdown error")
}

// TestServiceRegistry_GetAllServices ...
// Summary: TestServiceRegistry_GetAllServices
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return &mockUpstream{
				registerFunc: func(serviceName string) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
					serviceID, err := util.SanitizeServiceName(serviceName)
					require.NoError(t, err)
					return serviceID, nil, nil, nil
				},
			}, nil
		},
	}
	tm := &mockToolManager{}
	registry := New(f, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	// Initially, no services
	services, err := registry.GetAllServices()
	require.NoError(t, err)
	assert.Empty(t, services)

	// Register two services
	serviceConfig1 := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("service1"),
	}.Build()
	_, _, _, err = registry.RegisterService(context.Background(), serviceConfig1)
	require.NoError(t, err)

	serviceConfig2 := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("service2"),
	}.Build()
	_, _, _, err = registry.RegisterService(context.Background(), serviceConfig2)
	require.NoError(t, err)

	// Get all services
	services, err = registry.GetAllServices()
	require.NoError(t, err)
	assert.Len(t, services, 2)
}

// TestServiceRegistry_ServiceInfo ...
// Summary: TestServiceRegistry_ServiceInfo
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	registry := New(nil, nil, nil, nil, nil)

	// Get non-existent service info
	_, ok := registry.GetServiceInfo("non-existent")
	assert.False(t, ok)

	// Add and get service info
	serviceInfo := &tool.ServiceInfo{Name: "test-service"}
	registry.AddServiceInfo("service-id", serviceInfo)

	retrievedInfo, ok := registry.GetServiceInfo("service-id")
	require.True(t, ok)
	assert.Equal(t, serviceInfo, retrievedInfo)
}

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

type mockTool struct {
	tool *mcp_routerv1.Tool
}

// Tool ...
// Summary: Tool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return m.tool
}

// Execute ...
// Summary: Execute
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}

// GetCacheConfig ...
// Summary: GetCacheConfig
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

// MCPTool ...
// Summary: MCPTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t, _ := tool.ConvertProtoToMCPTool(m.tool)
	return t
}

type threadSafeToolManager struct {
	tool.ManagerInterface
	mu    sync.RWMutex
	tools map[string]tool.Tool
}

func newThreadSafeToolManager() *threadSafeToolManager {
	return &threadSafeToolManager{
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
	m.tools[t.Tool().GetName()] = t
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
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.tools[name]
	return t, ok
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
	m.mu.RLock()
	defer m.mu.RUnlock()
	tools := make([]tool.Tool, 0, len(m.tools))
	for _, t := range m.tools {
		tools = append(tools, t)
	}
	return tools
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
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, t := range m.tools {
		if t.Tool().GetServiceId() == serviceID {
			count++
		}
	}
	return count
}

// TestServiceRegistry_RegisterService_DuplicateNameDoesNotClearExisting ...
// Summary: TestServiceRegistry_RegisterService_DuplicateNameDoesNotClearExisting
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return &mockUpstream{
				registerFunc: func(serviceName string) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
					serviceID, err := util.SanitizeServiceName(serviceName)
					require.NoError(t, err)
					return serviceID, nil, nil, nil
				},
			}, nil
		},
	}
	tm := newThreadSafeToolManager()
	registry := New(f, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	// Register the first service with a tool
	serviceConfig1 := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
	}.Build()
	serviceID, _, _, err := registry.RegisterService(context.Background(), serviceConfig1)
	require.NoError(t, err, "First registration should succeed")

	// Add a tool to the service
	tool1 := &mockTool{tool: mcp_routerv1.Tool_builder{Name: proto.String("tool1"), ServiceId: proto.String(serviceID)}.Build()}
	err = tm.AddTool(tool1)
	require.NoError(t, err)

	// Verify the tool is there
	_, ok := tm.GetTool("tool1")
	assert.True(t, ok, "Tool should be present after first registration")

	// Attempt to register another service with the same name
	serviceConfig2 := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
	}.Build()
	_, _, _, err = registry.RegisterService(context.Background(), serviceConfig2)
	require.Error(t, err, "Second registration with the same name should fail")

	// Verify that the tool from the first service is still there
	_, ok = tm.GetTool("tool1")
	assert.True(t, ok, "Tool should still be present after failed duplicate registration")
}

// mockPrompt is a mock implementation of prompt.Prompt for testing.
type mockPrompt struct {
	p         *mcp.Prompt
	serviceID string
}

// Prompt ...
// Summary: Prompt
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return m.p
}

// Service ...
// Summary: Service
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return m.serviceID
}

// Definition ...
// Summary: Definition
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

// Get ...
// Summary: Get
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}

// mockResource is a mock implementation of resource.Resource for testing.
type mockResource struct {
	r         *mcp.Resource
	serviceID string
}

// Resource ...
// Summary: Resource
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return m.r
}

// Service ...
// Summary: Service
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return m.serviceID
}

// Read ...
// Summary: Read
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}

// Subscribe ...
// Summary: Subscribe
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

// TestServiceRegistry_UnregisterService_ClearsAllData ...
// Summary: TestServiceRegistry_UnregisterService_ClearsAllData
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return &mockUpstream{
				registerFunc: func(serviceName string) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
					serviceID, err := util.SanitizeServiceName(serviceName)
					require.NoError(t, err)
					return serviceID, nil, nil, nil
				},
			}, nil
		},
	}
	tm := newThreadSafeToolManager()
	pm := prompt.NewManager()
	rm := resource.NewManager()
	registry := New(f, tm, pm, rm, auth.NewManager())

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
	}.Build()
	serviceID, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err, "Registration should succeed")

	// Manually add items to the managers
	err = tm.AddTool(&mockTool{tool: mcp_routerv1.Tool_builder{Name: proto.String("test-tool"), ServiceId: proto.String(serviceID)}.Build()})
	require.NoError(t, err)
	pm.AddPrompt(&mockPrompt{serviceID: serviceID, p: &mcp.Prompt{Name: "test-prompt"}})
	rm.AddResource(&mockResource{serviceID: serviceID, r: &mcp.Resource{URI: "test-resource"}})

	// Verify that the data is there
	assert.NotEmpty(t, tm.ListTools())
	assert.NotEmpty(t, pm.ListPrompts())
	assert.NotEmpty(t, rm.ListResources())

	// Unregister the service
	err = registry.UnregisterService(context.Background(), serviceID)
	require.NoError(t, err, "Unregistration should succeed")

	// Verify that all data has been cleared
	assert.Empty(t, tm.ListTools(), "Tools should be cleared after unregistration")
	assert.Empty(t, pm.ListPrompts(), "Prompts should be cleared after unregistration")
	assert.Empty(t, rm.ListResources(), "Resources should be cleared after unregistration")
}

// TestServiceRegistry_UnregisterService_CallsShutdown ...
// Summary: TestServiceRegistry_UnregisterService_CallsShutdown
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	shutdownCalled := false
	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return &mockUpstream{
				registerFunc: func(serviceName string) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
					serviceID, err := util.SanitizeServiceName(serviceName)
					require.NoError(t, err)
					return serviceID, nil, nil, nil
				},
				shutdownFunc: func() error {
					shutdownCalled = true
					return nil
				},
			}, nil
		},
	}
	tm := newThreadSafeToolManager()
	registry := New(f, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
	}.Build()
	serviceID, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err, "Registration should succeed")

	// Unregister the service
	err = registry.UnregisterService(context.Background(), serviceID)
	require.NoError(t, err, "Unregistration should succeed")

	// Verify that the shutdown method was called
	assert.True(t, shutdownCalled, "Shutdown method should be called on unregister")
}

// TestGetServiceError ...
// Summary: TestGetServiceError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	registry := New(nil, nil, nil, nil, nil)

	// No error initially
	err, ok := registry.GetServiceError("test")
	assert.False(t, ok)
	assert.Equal(t, "", err)

	// Inject error manually (using internal map access via reflection or just simulating a fail state if possible)
	// Since we cannot access internal map easily from outside test package if it was external,
	// but we are in package serviceregistry!
	registry.serviceErrors["test"] = "some error"

	err, ok = registry.GetServiceError("test")
	assert.True(t, ok)
	assert.Equal(t, "some error", err)
}

// TestServiceRegistry_RegisterService_RetryFailed ...
// Summary: TestServiceRegistry_RegisterService_RetryFailed
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// First attempt fails
	failFactory := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return nil, errors.New("init fail")
		},
	}
	tm := &mockToolManager{}
	registry := New(failFactory, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("retry-service"),
	}.Build()
	_, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.Error(t, err)

	// Verify error state
	serviceID, _ := util.SanitizeServiceName("retry-service")
	msg, ok := registry.GetServiceError(serviceID)
	assert.True(t, ok)
	assert.Contains(t, msg, "init fail")

	// Second attempt succeeds
	successFactory := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return &mockUpstream{
				registerFunc: func(_ string) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
					return serviceID, nil, nil, nil
				},
			}, nil
		},
	}
	// Switch factory (hacky but we have reference)
	registry.factory = successFactory

	sID, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err)
	assert.Equal(t, serviceID, sID)

	// Verify error cleared
	msg, ok = registry.GetServiceError(serviceID)
	assert.False(t, ok)
}
