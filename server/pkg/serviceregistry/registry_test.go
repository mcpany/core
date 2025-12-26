// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package serviceregistry

import (
	"context"
	"errors"
	"testing"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream"
	"github.com/mcpany/core/pkg/upstream/factory"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mocks

type mockUpstream struct {
	upstream.Upstream
	registerFunc func(serviceName string) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error)
	shutdownFunc func() error
}

func (m *mockUpstream) Shutdown(_ context.Context) error {
	if m.shutdownFunc != nil {
		return m.shutdownFunc()
	}
	return nil
}

func (m *mockUpstream) Register(_ context.Context, serviceConfig *configv1.UpstreamServiceConfig, _ tool.ManagerInterface, _ prompt.ManagerInterface, _ resource.ManagerInterface, _ bool) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	if m.registerFunc != nil {
		return m.registerFunc(serviceConfig.GetName())
	}
	return "mock-service-key", nil, nil, nil
}

type mockFactory struct {
	factory.Factory
	newUpstreamFunc func() (upstream.Upstream, error)
}

func (m *mockFactory) NewUpstream(_ *configv1.UpstreamServiceConfig) (upstream.Upstream, error) {
	if m.newUpstreamFunc != nil {
		return m.newUpstreamFunc()
	}
	return &mockUpstream{}, nil
}

type mockToolManager struct {
	tool.ManagerInterface
}

func (m *mockToolManager) AddTool(_ tool.Tool) error             { return nil }
func (m *mockToolManager) ClearToolsForService(_ string)         {}
func (m *mockToolManager) GetTool(_ string) (tool.Tool, bool)    { return nil, false }
func (m *mockToolManager) ListTools() []tool.Tool                { return nil }
func (m *mockToolManager) ListServices() []*tool.ServiceInfo     { return nil }
func (m *mockToolManager) SetMCPServer(_ tool.MCPServerProvider) {}
func (m *mockToolManager) ExecuteTool(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return nil, nil
}
func (m *mockToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}

func TestNew(t *testing.T) {
	pm := pool.NewManager()
	f := factory.NewUpstreamServiceFactory(pm)
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

func TestServiceRegistry_RegisterAndGetService(t *testing.T) {
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

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-service")
	httpService := &configv1.HttpUpstreamService{}
	httpService.SetAddress("http://localhost")
	serviceConfig.SetHttpService(httpService)

	authConfig := &configv1.AuthenticationConfig{}
	apiKeyAuth := &configv1.APIKeyAuth{}
	apiKeyAuth.SetKeyValue("test-key")
	authConfig.SetApiKey(apiKeyAuth)
	serviceConfig.SetAuthentication(authConfig)

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
	assert.Equal(t, serviceConfig, retrievedConfig)

	// Check authenticator
	_, ok = am.GetAuthenticator(serviceID)
	assert.True(t, ok, "Authenticator should have been added")

	// Get non-existent config
	_, ok = registry.GetServiceConfig("non-existent")
	assert.False(t, ok)

	t.Run("with OAuth2 authenticator", func(t *testing.T) {
		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("oauth2-service")
		httpService := &configv1.HttpUpstreamService{}
		httpService.SetAddress("http://localhost")
		serviceConfig.SetHttpService(httpService)
		authConfig := &configv1.AuthenticationConfig{}
		oauth2Config := &configv1.OAuth2Auth{}
		oauth2Config.SetIssuerUrl("https://accounts.google.com")
		oauth2Config.SetAudience("test-audience")
		authConfig.SetOauth2(oauth2Config)
		serviceConfig.SetAuthentication(authConfig)
		serviceID, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
		require.NoError(t, err)

		_, ok := am.GetAuthenticator(serviceID)
		assert.True(t, ok, "OAuth2 authenticator should have been added")
	})
}

func TestServiceRegistry_RegisterService_FactoryError(t *testing.T) {
	factoryErr := errors.New("factory error")
	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return nil, factoryErr
		},
	}
	tm := &mockToolManager{}
	registry := New(f, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-service")
	_, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), factoryErr.Error())
}

func TestServiceRegistry_RegisterService_UpstreamError(t *testing.T) {
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

func TestServiceRegistry_RegisterService_DuplicateName(t *testing.T) {
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

	serviceConfig1 := &configv1.UpstreamServiceConfig{}
	serviceConfig1.SetName("test-service")
	_, _, _, err := registry.RegisterService(context.Background(), serviceConfig1)
	require.NoError(t, err, "First registration should succeed")

	// Attempt to register another service with the same name
	serviceConfig2 := &configv1.UpstreamServiceConfig{}
	serviceConfig2.SetName("test-service")
	_, _, _, err = registry.RegisterService(context.Background(), serviceConfig2)
	require.Error(t, err, "Second registration with the same name should fail")
	assert.Contains(t, err.Error(), `service with name "test-service" already registered`)
}

func TestServiceRegistry_UnregisterService(t *testing.T) {
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

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-service")

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
	assert.Contains(t, err.Error(), `service "non-existent-service" not found`)
}

func TestServiceRegistry_Close(t *testing.T) {
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
	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-service")
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
	serviceConfig2 := &configv1.UpstreamServiceConfig{}
	serviceConfig2.SetName("test-service-2")
	_, _, _, err = registry.RegisterService(context.Background(), serviceConfig2)
	require.NoError(t, err)

	err = registry.Close(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "shutdown error")
}

func TestServiceRegistry_GetAllServices(t *testing.T) {
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
	serviceConfig1 := &configv1.UpstreamServiceConfig{}
	serviceConfig1.SetName("service1")
	_, _, _, err = registry.RegisterService(context.Background(), serviceConfig1)
	require.NoError(t, err)

	serviceConfig2 := &configv1.UpstreamServiceConfig{}
	serviceConfig2.SetName("service2")
	_, _, _, err = registry.RegisterService(context.Background(), serviceConfig2)
	require.NoError(t, err)

	// Get all services
	services, err = registry.GetAllServices()
	require.NoError(t, err)
	assert.Len(t, services, 2)
}

func TestServiceRegistry_ServiceInfo(t *testing.T) {
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
