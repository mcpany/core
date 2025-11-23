// Copyright (C) 2025 Author(s) of MCP Any
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
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mocks

type mockUpstream struct {
	upstream.Upstream
	registerFunc func() (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error)
}

func (m *mockUpstream) Register(ctx context.Context, serviceConfig *configv1.UpstreamServiceConfig, toolManager tool.ToolManagerInterface, promptManager prompt.PromptManagerInterface, resourceManager resource.ResourceManagerInterface, isReload bool) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	if m.registerFunc != nil {
		return m.registerFunc()
	}
	return "mock-service-key", nil, nil, nil
}

type mockFactory struct {
	factory.Factory
	newUpstreamFunc func() (upstream.Upstream, error)
}

func (m *mockFactory) NewUpstream(config *configv1.UpstreamServiceConfig) (upstream.Upstream, error) {
	if m.newUpstreamFunc != nil {
		return m.newUpstreamFunc()
	}
	return &mockUpstream{}, nil
}

type mockToolManager struct {
	tool.ToolManagerInterface
}

func (m *mockToolManager) AddTool(t tool.Tool) error                               { return nil }
func (m *mockToolManager) ClearToolsForService(serviceID string)                   {}
func (m *mockToolManager) GetTool(name string) (tool.Tool, bool)                   { return nil, false }
func (m *mockToolManager) ListTools() []tool.Tool                                  { return nil }
func (m *mockToolManager) SetMCPServer(mcpServer tool.MCPServerProvider) {}
func (m *mockToolManager) ExecuteTool(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return nil, nil
}

func TestNew(t *testing.T) {
	pm := pool.NewManager()
	f := factory.NewUpstreamServiceFactory(pm)
	tm := &mockToolManager{}
	prm := prompt.NewPromptManager()
	rm := resource.NewResourceManager()
	am := auth.NewAuthManager()

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
	f := &mockFactory{}
	tm := &mockToolManager{}
	prm := prompt.NewPromptManager()
	rm := resource.NewResourceManager()
	am := auth.NewAuthManager()
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
	assert.Equal(t, "mock-service-key", serviceID)
	assert.Nil(t, tools)
	assert.Nil(t, resources)

	// Get the config
	retrievedConfig, ok := registry.GetServiceConfig("mock-service-key")
	require.True(t, ok)
	assert.Equal(t, serviceConfig, retrievedConfig)

	// Check authenticator
	_, ok = am.GetAuthenticator("mock-service-key")
	assert.True(t, ok, "Authenticator should have been added")

	// Get non-existent config
	_, ok = registry.GetServiceConfig("non-existent")
	assert.False(t, ok)
}

func TestServiceRegistry_RegisterService_FactoryError(t *testing.T) {
	factoryErr := errors.New("factory error")
	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return nil, factoryErr
		},
	}
	tm := &mockToolManager{}
	registry := New(f, tm, prompt.NewPromptManager(), resource.NewResourceManager(), auth.NewAuthManager())

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
				registerFunc: func() (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
					return "", nil, nil, upstreamErr
				},
			}, nil
		},
	}
	tm := &mockToolManager{}
	registry := New(f, tm, prompt.NewPromptManager(), resource.NewResourceManager(), auth.NewAuthManager())

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
				registerFunc: func() (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
					// Both services will have the same key because they have the same name
					return "test-service_e3b0c442", nil, nil, nil
				},
			}, nil
		},
	}
	tm := &mockToolManager{}
	registry := New(f, tm, prompt.NewPromptManager(), resource.NewResourceManager(), auth.NewAuthManager())

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
