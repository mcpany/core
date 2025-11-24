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

package serviceregistry

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockToolManagerWithCleanupTracking tracks calls to ClearToolsForService.
type mockToolManagerWithCleanupTracking struct {
	mockToolManager
	clearCalls map[string]int
	mu         sync.Mutex
}

func newMockToolManagerWithCleanupTracking() *mockToolManagerWithCleanupTracking {
	return &mockToolManagerWithCleanupTracking{
		clearCalls: make(map[string]int),
	}
}

func (m *mockToolManagerWithCleanupTracking) ClearToolsForService(serviceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clearCalls[serviceID]++
}

func (m *mockToolManagerWithCleanupTracking) getClearCallCount(serviceID string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.clearCalls[serviceID]
}

// mockResource is a simple implementation of the resource.Resource interface for testing.
type mockResource struct {
	resource *mcp.Resource
	service  string
}

func (m *mockResource) Resource() *mcp.Resource {
	return m.resource
}

func (m *mockResource) Service() string {
	return m.service
}

func (m *mockResource) Read(ctx context.Context) (*mcp.ReadResourceResult, error) {
	return nil, nil // Not needed for this test
}

func (m *mockResource) Subscribe(ctx context.Context) error {
	return nil // Not needed for this test
}

// mockPrompt is a simple implementation of the prompt.Prompt interface for testing.
type mockPrompt struct {
	prompt  *mcp.Prompt
	service string
}

func (m *mockPrompt) Prompt() *mcp.Prompt {
	return m.prompt
}

func (m *mockPrompt) Service() string {
	return m.service
}

func (m *mockPrompt) Get(ctx context.Context, args json.RawMessage) (*mcp.GetPromptResult, error) {
	return nil, nil // Not needed for this test
}

// realisticMockUpstream is a mock that interacts with managers during registration.
type realisticMockUpstream struct {
	upstream.Upstream
	serviceID   string
	resourceDef *configv1.ResourceDefinition
	promptDef   *configv1.PromptDefinition
}

func (m *realisticMockUpstream) Register(
	ctx context.Context,
	serviceConfig *configv1.UpstreamServiceConfig,
	toolManager tool.ToolManagerInterface,
	promptManager prompt.PromptManagerInterface,
	resourceManager resource.ResourceManagerInterface,
	isReload bool,
) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	// Simulate resource registration
	if m.resourceDef != nil {
		res := &mockResource{
			resource: &mcp.Resource{URI: fmt.Sprintf("mcp://%s/r/%s", m.serviceID, m.resourceDef.GetName())},
			service:  m.serviceID,
		}
		resourceManager.AddResource(res)
	}

	// Simulate prompt registration
	if m.promptDef != nil {
		p := &mockPrompt{
			prompt:  &mcp.Prompt{Name: fmt.Sprintf("%s/%s", m.serviceID, m.promptDef.GetName())},
			service: m.serviceID,
		}
		promptManager.AddPrompt(p)
	}

	// Return a dummy tool definition to simulate tool discovery
	toolDef := &configv1.ToolDefinition{}
	toolDef.SetName("test-tool")

	return m.serviceID, []*configv1.ToolDefinition{toolDef}, []*configv1.ResourceDefinition{m.resourceDef}, nil
}

// TestServiceRegistry_RegisterService_DuplicateName_Cleanup confirms that when a
// service registration fails due to a duplicate name, the cleanup process is
// correctly triggered for all managers.
func TestServiceRegistry_RegisterService_DuplicateName_Cleanup(t *testing.T) {
	serviceName := "duplicate-service"
	serviceID := "duplicate-service_e3b0c442" // Name with the empty hash
	resourceName := "test-resource"
	promptName := "test-prompt"

	// Resource and Prompt definitions
	resourceDef := &configv1.ResourceDefinition{}
	resourceDef.SetName(resourceName)

	promptDef := &configv1.PromptDefinition{}
	promptDef.SetName(promptName)

	// Mocks
	mockFactory := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return &realisticMockUpstream{
				serviceID:   serviceID,
				resourceDef: resourceDef,
				promptDef:   promptDef,
			}, nil
		},
	}
	mockToolManager := newMockToolManagerWithCleanupTracking()
	mockPromptManager := prompt.NewPromptManager()
	mockResourceManager := resource.NewResourceManager()

	registry := New(mockFactory, mockToolManager, mockPromptManager, mockResourceManager, auth.NewAuthManager())

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName(serviceName)

	// First registration (should succeed)
	_, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err, "First registration should succeed")

	resourceID := fmt.Sprintf("mcp://%s/r/%s", serviceID, resourceName)
	promptID := fmt.Sprintf("%s/%s", serviceID, promptName)

	// Verify that resources from the first registration are present
	_, ok := mockResourceManager.GetResource(resourceID)
	require.True(t, ok, "Resource from the first registration should be present")

	_, ok = mockPromptManager.GetPrompt(promptID)
	require.True(t, ok, "Prompt from the first registration should be present")

	// Second registration (should fail)
	_, _, _, err = registry.RegisterService(context.Background(), serviceConfig)
	require.Error(t, err, "Second registration should fail")

	// Verification of cleanup
	assert.Contains(t, err.Error(), "already registered", "Error message should indicate a duplicate service")
	assert.Equal(t, 1, mockToolManager.getClearCallCount(serviceID), "ClearToolsForService should be called once for the duplicate service")

	// This is the core of the bug: the test should fail here because resources
	// and prompts are not being cleaned up.
	_, ok = mockResourceManager.GetResource(resourceID)
	assert.False(t, ok, "Resource from the failed registration should be cleaned up")

	_, ok = mockPromptManager.GetPrompt(promptID)
	assert.False(t, ok, "Prompt from the failed registration should be cleaned up")
}

func TestServiceRegistry_AddAndGetServiceInfo(t *testing.T) {
	registry := New(nil, nil, nil, nil, nil)
	serviceID := "test-service"
	serviceInfo := &tool.ServiceInfo{
		Name: "Test service",
	}

	registry.AddServiceInfo(serviceID, serviceInfo)

	retrievedInfo, ok := registry.GetServiceInfo(serviceID)
	require.True(t, ok)
	assert.Equal(t, serviceInfo, retrievedInfo)

	_, ok = registry.GetServiceInfo("non-existent")
	assert.False(t, ok)
}

func TestServiceRegistry_UnregisterService(t *testing.T) {
	f := &mockFactory{}
	tm := newMockToolManagerWithCleanupTracking()
	prm := prompt.NewPromptManager()
	rm := resource.NewResourceManager()
	am := auth.NewAuthManager()
	registry := New(f, tm, prm, rm, am)

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-service")
	serviceID := "test-service_e3b0c442"

	registry.serviceConfigs[serviceID] = serviceConfig
	registry.AddServiceInfo(serviceID, &tool.ServiceInfo{})
	am.AddAuthenticator(serviceID, &auth.APIKeyAuthenticator{})

	err := registry.UnregisterService(context.Background(), serviceID)
	require.NoError(t, err)

	_, ok := registry.GetServiceConfig(serviceID)
	assert.False(t, ok, "Service config should be removed")

	_, ok = registry.GetServiceInfo(serviceID)
	assert.False(t, ok, "Service info should be removed")

	assert.Equal(t, 1, tm.getClearCallCount(serviceID), "ClearToolsForService should be called")

	_, ok = am.GetAuthenticator(serviceID)
	assert.False(t, ok, "Authenticator should be removed")

	err = registry.UnregisterService(context.Background(), "non-existent")
	require.Error(t, err)
}

func TestServiceRegistry_GetAllServices(t *testing.T) {
	registry := New(nil, nil, nil, nil, nil)
	serviceConfig1 := &configv1.UpstreamServiceConfig{}
	serviceConfig1.SetName("service1")
	serviceConfig2 := &configv1.UpstreamServiceConfig{}
	serviceConfig2.SetName("service2")

	registry.serviceConfigs["service1"] = serviceConfig1
	registry.serviceConfigs["service2"] = serviceConfig2

	services, err := registry.GetAllServices()
	require.NoError(t, err)
	assert.Len(t, services, 2)
	assert.Contains(t, services, serviceConfig1)
	assert.Contains(t, services, serviceConfig2)
}
