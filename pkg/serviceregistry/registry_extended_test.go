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

// mockToolManagerWithCleanupTracking tracks calls to PruneToolsForService.
type mockToolManagerWithCleanupTracking struct {
	mockToolManager
	pruneCalls map[string]int
	mu         sync.Mutex
}

func newMockToolManagerWithCleanupTracking() *mockToolManagerWithCleanupTracking {
	return &mockToolManagerWithCleanupTracking{
		pruneCalls: make(map[string]int),
	}
}

func (m *mockToolManagerWithCleanupTracking) PruneToolsForService(serviceID string, keepToolNames []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pruneCalls[serviceID]++

	// Basic prune logic for test
	keepKeys := make(map[string]struct{})
	for _, name := range keepToolNames {
		// Assume sanitized name for mock
		// The mockToolManager uses serviceID + "." + sanitizedToolName
		// But here we might not have util.SanitizeToolName handy or it might be complex
		// Let's iterate and check suffix? Or just trust test setup uses simple names
		key := serviceID + "." + name // Assuming name is already what is stored or simple enough
		keepKeys[key] = struct{}{}
	}
	// For this test, mockUpstream returns "test-tool" and "test-tool-2".
	// The mockToolManager AddTool does sanitization.
	// So we should try to match that.
	// But `PruneToolsForService` implementation in `ToolManager` sanitizes `keepToolNames`.
	// Here in mock we can just clear everything NOT in keep list.
	// But `keepToolNames` are original names.
	// We'll skip complex logic and just rely on `pruneCalls` count for this specific mock behavior,
	// unless we want to verify tools are gone.
	// Let's implement full logic if possible.
}

func (m *mockToolManagerWithCleanupTracking) getPruneCallCount(serviceID string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.pruneCalls[serviceID]
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

// TestServiceRegistry_RegisterService_Overwrite_Pruning confirms that when a
// service is re-registered (overwrite), the pruning process is correctly triggered
// for all managers to remove stale items.
func TestServiceRegistry_RegisterService_Overwrite_Pruning(t *testing.T) {
	serviceName := "duplicate-service"
	serviceID := "duplicate-service_e3b0c442" // Name with the empty hash

	// Initial set of items
	resourceName1 := "resource-1"
	promptName1 := "prompt-1"
	resourceDef1 := &configv1.ResourceDefinition{}
	resourceDef1.SetName(resourceName1)
	resourceDef1.SetUri(fmt.Sprintf("mcp://%s/r/%s", serviceID, resourceName1))
	promptDef1 := &configv1.PromptDefinition{}
	promptDef1.SetName(promptName1)

	// New set of items (replacing old ones)
	resourceName2 := "resource-2"
	promptName2 := "prompt-2"
	resourceDef2 := &configv1.ResourceDefinition{}
	resourceDef2.SetName(resourceName2)
	resourceDef2.SetUri(fmt.Sprintf("mcp://%s/r/%s", serviceID, resourceName2))
	promptDef2 := &configv1.PromptDefinition{}
	promptDef2.SetName(promptName2)

	// Mocks
	upstream1 := &realisticMockUpstream{
		serviceID:   serviceID,
		resourceDef: resourceDef1,
		promptDef:   promptDef1,
	}
	upstream2 := &realisticMockUpstream{
		serviceID:   serviceID,
		resourceDef: resourceDef2,
		promptDef:   promptDef2,
	}

	mockFactory := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return upstream1, nil
		},
	}
	mockToolManager := newMockToolManagerWithCleanupTracking()
	mockPromptManager := prompt.NewPromptManager()
	mockResourceManager := resource.NewResourceManager()

	registry := New(mockFactory, mockToolManager, mockPromptManager, mockResourceManager, auth.NewAuthManager())

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName(serviceName)

	// First registration
	_, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err, "First registration should succeed")

	resourceID1 := fmt.Sprintf("mcp://%s/r/%s", serviceID, resourceName1)
	promptID1 := fmt.Sprintf("%s/%s", serviceID, promptName1)

	// Verify first items are present
	_, ok := mockResourceManager.GetResource(resourceID1)
	require.True(t, ok, "Resource 1 should be present")
	_, ok = mockPromptManager.GetPrompt(promptID1)
	require.True(t, ok, "Prompt 1 should be present")

	// Update factory to return second upstream
	mockFactory.newUpstreamFunc = func() (upstream.Upstream, error) {
		return upstream2, nil
	}

	// Second registration (overwrite)
	_, _, _, err = registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err, "Second registration should succeed (overwrite)")

	// Verify prune call count
	assert.Equal(t, 1, mockToolManager.getPruneCallCount(serviceID), "PruneToolsForService should be called once")

	// Verify first items are GONE
	_, ok = mockResourceManager.GetResource(resourceID1)
	assert.False(t, ok, "Resource 1 should be pruned")
	_, ok = mockPromptManager.GetPrompt(promptID1)
	assert.False(t, ok, "Prompt 1 should be pruned")

	// Verify second items are PRESENT
	resourceID2 := fmt.Sprintf("mcp://%s/r/%s", serviceID, resourceName2)
	promptID2 := fmt.Sprintf("%s/%s", serviceID, promptName2)

	_, ok = mockResourceManager.GetResource(resourceID2)
	assert.True(t, ok, "Resource 2 should be present")
	_, ok = mockPromptManager.GetPrompt(promptID2)
	assert.True(t, ok, "Prompt 2 should be present")
}
