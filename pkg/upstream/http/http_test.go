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

package http

import (
	"context"
	"errors"
	"testing"

	"github.com/mcpany/core/pkg/client"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestHttpMethodToString(t *testing.T) {
	testCases := []struct {
		name          string
		method        configv1.HttpCallDefinition_HttpMethod
		expected      string
		expectAnError bool
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
			name:          "unspecified",
			method:        configv1.HttpCallDefinition_HTTP_METHOD_UNSPECIFIED,
			expected:      "",
			expectAnError: true,
		},
		{
			name:          "invalid",
			method:        configv1.HttpCallDefinition_HttpMethod(999),
			expected:      "",
			expectAnError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := httpMethodToString(tc.method)
			if tc.expectAnError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestHTTPUpstream_Register(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewToolManager(nil)
		upstream := NewHTTPUpstream(pm)

		configJSON := `{"name": "test-service", "http_service": {"address": "http://localhost", "calls": [{"schema": {"name": "test-op"}, "method": "HTTP_METHOD_GET", "endpoint_path": "/test"}]}}`
		serviceConfig := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		serviceID, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.NoError(t, err)
		expectedKey, _ := util.SanitizeServiceName("test-service")
		assert.Equal(t, expectedKey, serviceID)
		assert.Len(t, discoveredTools, 1)
		p, ok := pool.Get[*client.HttpClientWrapper](pm, serviceID)
		assert.True(t, ok)
		assert.NotNil(t, p)
	})

	t.Run("nil http service config", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewToolManager(nil)
		upstream := NewHTTPUpstream(pm)

		configJSON := `{"name": "test-service", "grpc_service": {}}`
		serviceConfig := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "http service config is nil")
	})

	t.Run("invalid service name", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewToolManager(nil)
		upstream := NewHTTPUpstream(pm)

		configJSON := `{"name": "", "http_service": {}}`
		serviceConfig := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "id cannot be empty")
	})

	t.Run("tool registration with fallback operation ID", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewToolManager(nil)
		upstream := NewHTTPUpstream(pm)

		configJSON := `{"name": "test-service-fallback", "http_service": {"address": "http://localhost", "calls": [{"schema": {"description": "A test operation"}, "method": "HTTP_METHOD_GET", "endpoint_path": "/test"}, {"schema": {"description": ""}, "method": "HTTP_METHOD_POST", "endpoint_path": "/test2"}]}}`
		serviceConfig := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		serviceID, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		require.NoError(t, err)
		require.Len(t, discoveredTools, 2)

		tools := tm.ListTools()
		assert.Len(t, tools, 2)

		// Check for sanitized description as name
		sanitizedName := util.SanitizeOperationID("A test operation")
		sanitizedName, _ = util.SanitizeToolName(sanitizedName)
		toolID1 := serviceID + "." + sanitizedName
		_, ok := tm.GetTool(toolID1)
		assert.True(t, ok, "Tool with sanitized description should be found, expected %s", toolID1)

		// Check for default fallback name
		sanitizedName2, _ := util.SanitizeToolName("op_1")
		toolID2 := serviceID + "." + sanitizedName2
		_, ok = tm.GetTool(toolID2)
		assert.True(t, ok, "Tool with default fallback name should be found, expected %s", toolID2)
	})

	t.Run("authenticator creation fails", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewToolManager(nil)
		upstream := NewHTTPUpstream(pm)

		configJSON := `{"name": "auth-fail-service", "http_service": {"address": "http://localhost", "calls": [{"schema": {"name": "test-op"}, "method": "HTTP_METHOD_GET"}]}, "upstream_authentication": {"api_key": {}}}`
		serviceConfig := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		require.NoError(t, err)
		assert.Len(t, discoveredTools, 1, "Tool should still be registered even if auth is nil")
	})

	t.Run("registration with connection pool config", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewToolManager(nil)
		upstream := NewHTTPUpstream(pm)

		configJSON := `{"name": "test-service-with-pool", "http_service": {"address": "http://localhost", "calls": [{"schema": {"name": "test-op"}, "method": "HTTP_METHOD_GET", "endpoint_path": "/test"}]}, "connection_pool": {"max_connections": 50, "max_idle_connections": 10}}`
		serviceConfig := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		// We need to replace the NewHttpPool function with a mock to check the parameters.
		originalNewHttpPool := NewHttpPool
		defer func() { NewHttpPool = originalNewHttpPool }()

		var capturedMinSize, capturedMaxSize, capturedIdleTimeout int
		NewHttpPool = func(minSize, maxSize, idleTimeout int, config *configv1.UpstreamServiceConfig) (pool.Pool[*client.HttpClientWrapper], error) {
			capturedMinSize = minSize
			capturedMaxSize = maxSize
			capturedIdleTimeout = idleTimeout
			return originalNewHttpPool(minSize, maxSize, idleTimeout, config)
		}

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.NoError(t, err)

		assert.Equal(t, 10, capturedMinSize)
		assert.Equal(t, 50, capturedMaxSize)
		assert.Equal(t, 300, capturedIdleTimeout)
	})
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

	configJSON := `{"name": "add-tool-fail-service", "http_service": {"address": "http://localhost", "calls": [{"schema": {"name": "test-op"}, "method": "HTTP_METHOD_GET"}]}}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	var promptManager prompt.PromptManagerInterface
	var resourceManager resource.ResourceManagerInterface

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, mockTm, promptManager, resourceManager, false)

	require.NoError(t, err)
	assert.Empty(t, discoveredTools, "No tools should be discovered if AddTool fails")
	assert.Empty(t, mockTm.ListTools(), "Tool manager should be empty if AddTool fails")
}

func TestHTTPUpstream_Register_WithReload(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewToolManager(nil)
	upstream := NewHTTPUpstream(pm)

	// Initial registration
	configJSON1 := `{"name": "reload-test", "http_service": {"address": "http://localhost", "calls": [{"schema": {"name": "op1"}, "method": "HTTP_METHOD_GET"}]}}`
	serviceConfig1 := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON1), serviceConfig1))

	serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig1, tm, nil, nil, false)
	require.NoError(t, err)
	assert.Len(t, tm.ListTools(), 1)

	// Reload with a different tool
	configJSON2 := `{"name": "reload-test", "http_service": {"address": "http://localhost", "calls": [{"schema": {"name": "op2"}, "method": "HTTP_METHOD_GET"}]}}`
	serviceConfig2 := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON2), serviceConfig2))

	// We need to use a mock tool manager to check if ClearToolsForService is called.
	// Since the upstream doesn't expose the tool manager, this is the best we can do
	// without further refactoring. The logic is tested implicitly by checking the
	// final state of the tools.
	_, _, _, err = upstream.Register(context.Background(), serviceConfig2, tm, nil, nil, true)
	require.NoError(t, err)
	assert.Len(t, tm.ListTools(), 1)
	sanitizedToolName, _ := util.SanitizeToolName("op2")
	toolID2 := serviceID + "." + sanitizedToolName
	_, ok := tm.GetTool(toolID2)
	assert.True(t, ok)
	sanitizedToolName, _ = util.SanitizeToolName("op1")
	toolID1 := serviceID + "." + sanitizedToolName
	_, ok = tm.GetTool(toolID1)
	assert.False(t, ok)
}

func TestHTTPUpstream_Register_InvalidMethod(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewToolManager(nil)
	upstream := NewHTTPUpstream(pm)

	configJSON := `{"name": "test-service-invalid-method", "http_service": {"address": "http://localhost", "calls": [{"schema": {"name": "test-op"}, "method": 999, "endpoint_path": "/test"}]}}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	assert.NoError(t, err)
	assert.Len(t, discoveredTools, 0, "No tools should be registered for an invalid method")
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
		sanitizedToolName, _ := util.SanitizeToolName(t.Tool().GetName())
		toolID := t.Tool().GetServiceId() + "." + sanitizedToolName
		if toolID == name {
			return t, true
		}
	}
	return nil, false
}

func (m *mockToolManager) ListTools() []tool.Tool {
	return m.addedTools
}

func (m *mockToolManager) ClearToolsForService(serviceID string) {
	if m.failOnClear {
		// To test error handling if clearing was to fail, although the
		// current implementation does not return an error.
		return
	}
	var remainingTools []tool.Tool
	for _, t := range m.addedTools {
		if t.Tool().GetServiceId() != serviceID {
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

func TestHTTPUpstream_URLConstruction(t *testing.T) {
	testCases := []struct {
		name          string
		address       string
		endpointPath  string
		expectedFqn   string
		expectAnError bool
	}{
		{
			name:         "trailing slash in address",
			address:      "http://localhost:8080/",
			endpointPath: "api/v1/test",
			expectedFqn:  "GET http://localhost:8080/api/v1/test",
		},
		{
			name:         "leading slash in endpoint",
			address:      "http://localhost:8080",
			endpointPath: "/api/v1/test",
			expectedFqn:  "GET http://localhost:8080/api/v1/test",
		},
		{
			name:         "both slashes present",
			address:      "http://localhost:8080/",
			endpointPath: "/api/v1/test",
			expectedFqn:  "GET http://localhost:8080/api/v1/test",
		},
		{
			name:         "no slashes",
			address:      "http://localhost:8080",
			endpointPath: "api/v1/test",
			expectedFqn:  "GET http://localhost:8080/api/v1/test",
		},
		{
			name:         "address with scheme",
			address:      "https://api.example.com",
			endpointPath: "/users",
			expectedFqn:  "GET https://api.example.com/users",
		},
		{
			name:         "double slash bug",
			address:      "http://localhost:8080/",
			endpointPath: "/api/v1/test",
			expectedFqn:  "GET http://localhost:8080/api/v1/test",
		},
		{
			name:         "address with path and trailing slash",
			address:      "http://localhost:8080/base/",
			endpointPath: "api/v1/test",
			expectedFqn:  "GET http://localhost:8080/base/api/v1/test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pm := pool.NewManager()
			tm := tool.NewToolManager(nil)
			upstream := NewHTTPUpstream(pm)

			configJSON := `{"name": "url-test-service", "http_service": {"address": "` + tc.address + `", "calls": [{"schema": {"name": "test-op"}, "method": "HTTP_METHOD_GET", "endpoint_path": "` + tc.endpointPath + `"}]}}`
			serviceConfig := &configv1.UpstreamServiceConfig{}
			require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

			serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
			assert.NoError(t, err)

			sanitizedToolName, _ := util.SanitizeToolName("test-op")
			toolID := serviceID + "." + sanitizedToolName
			registeredTool, ok := tm.GetTool(toolID)
			assert.True(t, ok)
			assert.NotNil(t, registeredTool)
			assert.Equal(t, tc.expectedFqn, registeredTool.Tool().GetUnderlyingMethodFqn())
		})
	}
}
