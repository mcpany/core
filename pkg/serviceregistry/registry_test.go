/*
 * Copyright 2025 Author(s) of MCPXY
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

package serviceregistry

import (
	"context"
	"errors"
	"testing"

	"github.com/mcpxy/core/pkg/auth"
	"github.com/mcpxy/core/pkg/pool"
	"github.com/mcpxy/core/pkg/prompt"
	"github.com/mcpxy/core/pkg/resource"
	"github.com/mcpxy/core/pkg/tool"
	"github.com/mcpxy/core/pkg/upstream"
	"github.com/mcpxy/core/pkg/upstream/factory"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mocks

type mockUpstream struct {
	upstream.Upstream
	registerFunc func() (string, []*configv1.ToolDefinition, error)
}

func (m *mockUpstream) Register(ctx context.Context, serviceConfig *configv1.UpstreamServiceConfig, toolManager tool.ToolManagerInterface, promptManager prompt.PromptManagerInterface, resourceManager resource.ResourceManagerInterface, isReload bool) (string, []*configv1.ToolDefinition, error) {
	if m.registerFunc != nil {
		return m.registerFunc()
	}
	return "mock-service-key", nil, nil
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
func (m *mockToolManager) ClearToolsForService(serviceKey string)                  {}
func (m *mockToolManager) GetTool(name string) (tool.Tool, bool)                   { return nil, false }
func (m *mockToolManager) ListTools() []tool.Tool                                  { return nil }
func (m *mockToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {}
func (m *mockToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	return nil, false
}
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
	serviceKey, tools, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err)
	assert.Equal(t, "mock-service-key", serviceKey)
	assert.Nil(t, tools)

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
	registry := New(f, &mockToolManager{}, prompt.NewPromptManager(), resource.NewResourceManager(), auth.NewAuthManager())

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-service")
	_, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), factoryErr.Error())
}

func TestServiceRegistry_RegisterService_UpstreamError(t *testing.T) {
	upstreamErr := errors.New("upstream error")
	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return &mockUpstream{
				registerFunc: func() (string, []*configv1.ToolDefinition, error) {
					return "", nil, upstreamErr
				},
			}, nil
		},
	}
	registry := New(f, &mockToolManager{}, prompt.NewPromptManager(), resource.NewResourceManager(), auth.NewAuthManager())

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-service")
	_, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), upstreamErr.Error())
}
