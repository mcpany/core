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

package http

import (
	"context"
	"errors"
	"testing"

	"github.com/mcpxy/core/pkg/client"
	"github.com/mcpxy/core/pkg/pool"
	"github.com/mcpxy/core/pkg/prompt"
	"github.com/mcpxy/core/pkg/resource"
	"github.com/mcpxy/core/pkg/tool"
	"github.com/mcpxy/core/pkg/util"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHttpMethodToString(t *testing.T) {
	testCases := []struct {
		name     string
		method   configv1.HttpCallDefinition_HttpMethod
		expected string
	}{
		{
			name:     "GET",
			method:   configv1.HttpCallDefinition_HTTP_METHOD_GET,
			expected: "GET",
		},
		{
			name:     "POST",
			method:   configv1.HttpCallDefinition_HTTP_METHOD_POST,
			expected: "POST",
		},
		{
			name:     "PUT",
			method:   configv1.HttpCallDefinition_HTTP_METHOD_PUT,
			expected: "PUT",
		},
		{
			name:     "DELETE",
			method:   configv1.HttpCallDefinition_HTTP_METHOD_DELETE,
			expected: "DELETE",
		},
		{
			name:     "PATCH",
			method:   configv1.HttpCallDefinition_HTTP_METHOD_PATCH,
			expected: "PATCH",
		},
		{
			name:     "unspecified",
			method:   configv1.HttpCallDefinition_HTTP_METHOD_UNSPECIFIED,
			expected: "",
		},
		{
			name:     "invalid",
			method:   configv1.HttpCallDefinition_HttpMethod(999),
			expected: "",
		},
		{
			name:     "default",
			method:   configv1.HttpCallDefinition_HttpMethod(1000),
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := httpMethodToString(tc.method)
			if actual != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, actual)
			}
		})
	}
}

func TestHTTPUpstream_Register(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewToolManager()
		upstream := NewHTTPUpstream(pm)

		httpService := &configv1.HttpUpstreamService{}
		httpService.SetAddress("http://localhost")
		callDef := &configv1.HttpCallDefinition{}
		callDef.SetOperationId("test-op")
		callDef.SetMethod(configv1.HttpCallDefinition_HTTP_METHOD_GET)
		callDef.SetEndpointPath("/test")
		httpService.SetCalls([]*configv1.HttpCallDefinition{callDef})

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-service")
		serviceConfig.SetHttpService(httpService)

		serviceKey, discoveredTools, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.NoError(t, err)
		assert.Equal(t, "test-service", serviceKey)
		assert.Len(t, discoveredTools, 1)
		p, ok := pool.Get[*client.HttpClientWrapper](pm, serviceKey)
		assert.True(t, ok)
		assert.NotNil(t, p)
	})

	t.Run("nil http service config", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewToolManager()
		upstream := NewHTTPUpstream(pm)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-service")
		serviceConfig.SetGrpcService(&configv1.GrpcUpstreamService{})

		_, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "http service config is nil")
	})

	t.Run("invalid service name", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewToolManager()
		upstream := NewHTTPUpstream(pm)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("invalid name") // Invalid name with spaces

		_, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.Error(t, err)
	})

	t.Run("tool registration with fallback operation ID", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewToolManager()
		upstream := NewHTTPUpstream(pm)

		callDef1 := &configv1.HttpCallDefinition{}
		callDef1.SetDescription("A test operation")
		callDef1.SetMethod(configv1.HttpCallDefinition_HTTP_METHOD_GET)
		callDef1.SetEndpointPath("/test")

		callDef2 := &configv1.HttpCallDefinition{}
		callDef2.SetDescription("") // Empty description
		callDef2.SetMethod(configv1.HttpCallDefinition_HTTP_METHOD_POST)
		callDef2.SetEndpointPath("/test2")

		httpService := &configv1.HttpUpstreamService{}
		httpService.SetAddress("http://localhost")
		httpService.SetCalls([]*configv1.HttpCallDefinition{callDef1, callDef2})

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-service-fallback")
		serviceConfig.SetHttpService(httpService)

		_, discoveredTools, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		require.NoError(t, err)
		require.Len(t, discoveredTools, 2)

		tools := tm.ListTools()
		assert.Len(t, tools, 2)

		// Check for sanitized description as name
		sanitizedName := util.SanitizeOperationID("A test operation")
		toolID1 := "test-service-fallback/-/" + sanitizedName
		_, ok := tm.GetTool(toolID1)
		assert.True(t, ok, "Tool with sanitized description should be found, expected %s", toolID1)

		// Check for default fallback name
		toolID2 := "test-service-fallback/-/op1"
		_, ok = tm.GetTool(toolID2)
		assert.True(t, ok, "Tool with default fallback name should be found, expected %s", toolID2)
	})

	t.Run("authenticator creation fails", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewToolManager()
		upstream := NewHTTPUpstream(pm)

		callDef := &configv1.HttpCallDefinition{}
		callDef.SetOperationId("test-op")

		httpService := &configv1.HttpUpstreamService{}
		httpService.SetAddress("http://localhost")
		httpService.SetCalls([]*configv1.HttpCallDefinition{callDef})

		authConfig := &configv1.UpstreamAuthentication{}
		apiKeyAuth := &configv1.UpstreamAPIKeyAuth{}
		// An empty APIKeyAuth will cause NewUpstreamAuthenticator to return nil, which is handled gracefully.
		authConfig.SetApiKey(apiKeyAuth)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("auth-fail-service")
		serviceConfig.SetHttpService(httpService)
		serviceConfig.SetUpstreamAuthentication(authConfig)

		_, discoveredTools, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		require.NoError(t, err)
		assert.Len(t, discoveredTools, 1, "Tool should still be registered even if auth is nil")
	})
}

// mockToolManager to simulate errors
type mockToolManager struct {
	tool.ToolManagerInterface
	addError    error
	addedTools  []tool.Tool
	failOnClear bool
}

func (m *mockToolManager) AddTool(t tool.Tool) error {
	if m.addError != nil {
		return m.addError
	}
	m.addedTools = append(m.addedTools, t)
	return nil
}

func (m *mockToolManager) GetTool(name string) (tool.Tool, bool) {
	for _, t := range m.addedTools {
		// Simplified check: In a real scenario, you'd parse the name
		// and check against the tool's actual name and service key.
		// For this mock, we'll just assume the test provides the full tool ID.
		toolID, _ := util.GenerateToolID(t.Tool().GetServiceId(), t.Tool().GetName())
		if toolID == name {
			return t, true
		}
	}
	return nil, false
}

func (m *mockToolManager) ListTools() []tool.Tool {
	return m.addedTools
}

func (m *mockToolManager) ClearToolsForService(serviceKey string) {
	if m.failOnClear {
		// To test error handling if clearing was to fail, although the
		// current implementation does not return an error.
		return
	}
	var remainingTools []tool.Tool
	for _, t := range m.addedTools {
		if t.Tool().GetServiceId() != serviceKey {
			remainingTools = append(remainingTools, t)
		}
	}
	m.addedTools = remainingTools
}

func (m *mockToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {}
func (m *mockToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	return nil, false
}
func (m *mockToolManager) SetMCPServer(mcpServer tool.MCPServerProvider) {}
func (m *mockToolManager) ExecuteTool(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return nil, errors.New("not implemented")
}

func newMockToolManager() *mockToolManager {
	return &mockToolManager{
		addedTools: make([]tool.Tool, 0),
	}
}

func TestCreateAndRegisterHTTPTools_AddToolError(t *testing.T) {
	pm := pool.NewManager()
	mockTm := newMockToolManager()
	mockTm.addError = errors.New("failed to add tool")

	upstream := &HTTPUpstream{poolManager: pm}

	callDef := &configv1.HttpCallDefinition{}
	callDef.SetOperationId("test-op")

	httpService := &configv1.HttpUpstreamService{}
	httpService.SetAddress("http://localhost")
	httpService.SetCalls([]*configv1.HttpCallDefinition{callDef})

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("add-tool-fail-service")
	serviceConfig.SetHttpService(httpService)

	var promptManager prompt.PromptManagerInterface
	var resourceManager resource.ResourceManagerInterface

	_, discoveredTools, err := upstream.Register(context.Background(), serviceConfig, mockTm, promptManager, resourceManager, false)

	require.NoError(t, err)
	assert.Empty(t, discoveredTools, "No tools should be discovered if AddTool fails")
	assert.Empty(t, mockTm.ListTools(), "Tool manager should be empty if AddTool fails")
}

func TestHTTPUpstream_Register_WithReload(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewToolManager()
	upstream := NewHTTPUpstream(pm)

	// Initial registration
	httpService1 := &configv1.HttpUpstreamService{}
	httpService1.SetAddress("http://localhost")
	callDef1 := &configv1.HttpCallDefinition{}
	callDef1.SetOperationId("op1")
	httpService1.SetCalls([]*configv1.HttpCallDefinition{callDef1})
	serviceConfig1 := &configv1.UpstreamServiceConfig{}
	serviceConfig1.SetName("reload-test")
	serviceConfig1.SetHttpService(httpService1)

	_, _, err := upstream.Register(context.Background(), serviceConfig1, tm, nil, nil, false)
	require.NoError(t, err)
	assert.Len(t, tm.ListTools(), 1)

	// Reload with a different tool
	httpService2 := &configv1.HttpUpstreamService{}
	httpService2.SetAddress("http://localhost")
	callDef2 := &configv1.HttpCallDefinition{}
	callDef2.SetOperationId("op2")
	httpService2.SetCalls([]*configv1.HttpCallDefinition{callDef2})
	serviceConfig2 := &configv1.UpstreamServiceConfig{}
	serviceConfig2.SetName("reload-test")
	serviceConfig2.SetHttpService(httpService2)

	// We need to use a mock tool manager to check if ClearToolsForService is called.
	// Since the upstream doesn't expose the tool manager, this is the best we can do
	// without further refactoring. The logic is tested implicitly by checking the
	// final state of the tools.
	_, _, err = upstream.Register(context.Background(), serviceConfig2, tm, nil, nil, true)
	require.NoError(t, err)
	assert.Len(t, tm.ListTools(), 1)
	_, ok := tm.GetTool("reload-test/-/op2")
	assert.True(t, ok)
	_, ok = tm.GetTool("reload-test/-/op1")
	assert.False(t, ok)
}
