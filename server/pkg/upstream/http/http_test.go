// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http //nolint:revive,nolintlint // Package name 'http' is intentional for this directory structure.
import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
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

func TestHTTPUpstream_Register_InsecureSkipVerify(t *testing.T) {
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")

	// Create a test HTTPS server with a self-signed certificate
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("OK")) // This implicitly sets status to 200 OK if not already set
	}))
	defer server.Close()

	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "insecure-test",
		"http_service": {
			"address": "` + server.URL + `",
			"tools": [{"name": "test-op", "call_id": "test-op-call"}],
			"calls": {
				"test-op-call": {
					"id": "test-op-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/"
				}
			},
			"tls_config": {
				"insecure_skip_verify": true
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	// Verify that the tool can be called successfully, which proves the TLS handshake worked.
	sanitizedToolName, _ := util.SanitizeToolName("test-op")
	toolID := serviceID + "." + sanitizedToolName
	registeredTool, ok := tm.GetTool(toolID)
	require.True(t, ok)

	_, err = registeredTool.Execute(context.Background(), &tool.ExecutionRequest{})
	require.NoError(t, err)
}

func TestHTTPUpstream_Register_Disabled(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "disabled-tool-test",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [
				{"name": "op1", "call_id": "op1-call", "disable": true},
				{"name": "op2", "call_id": "op2-call", "disable": false}
			],
			"calls": {
				"op1-call": {
					"id": "op1-call",
					"method": "HTTP_METHOD_GET"
				},
				"op2-call": {
					"id": "op2-call",
					"method": "HTTP_METHOD_GET"
				}
			},
			"prompts": [
				{"name": "prompt1", "disable": true},
				{"name": "prompt2", "disable": false}
			],
			"resources": [
				{
					"name": "res1",
					"uri": "http://test1",
					"disable": true,
					"dynamic": {
						"http_call": {"id": "op2-call"}
					}
				},
				{
					"name": "res2",
					"uri": "http://test2",
					"disable": false,
					"dynamic": {
						"http_call": {"id": "op2-call"}
					}
				}
			]
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()

	serviceID, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
	require.NoError(t, err)

	// Check Tools
	assert.Len(t, discoveredTools, 1)
	assert.Equal(t, "op2", discoveredTools[0].GetName())

	sanitizedToolName1, _ := util.SanitizeToolName("op1")
	toolID1 := serviceID + "." + sanitizedToolName1
	_, ok := tm.GetTool(toolID1)
	assert.False(t, ok, "Disabled tool should not be registered")

	sanitizedToolName2, _ := util.SanitizeToolName("op2")
	toolID2 := serviceID + "." + sanitizedToolName2
	_, ok = tm.GetTool(toolID2)
	assert.True(t, ok, "Enabled tool should be registered")

	// Check Prompts
	sanitizedPromptName1, _ := util.SanitizeToolName("prompt1")
	promptID1 := serviceID + "." + sanitizedPromptName1
	_, ok = promptManager.GetPrompt(promptID1)
	assert.False(t, ok, "Disabled prompt should not be registered")

	sanitizedPromptName2, _ := util.SanitizeToolName("prompt2")
	promptID2 := serviceID + "." + sanitizedPromptName2
	_, ok = promptManager.GetPrompt(promptID2)
	assert.True(t, ok, "Enabled prompt should be registered")

	// Check Resources
	_, ok = resourceManager.GetResource("http://test1")
	assert.False(t, ok, "Disabled resource should not be registered")
	_, ok = resourceManager.GetResource("http://test2")
	assert.True(t, ok, "Enabled resource should be registered")
}

func TestDeterminismInToolNaming(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "test-determinism",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [
				{"call_id": "call2"},
				{"call_id": "call1"}
			],
			"calls": {
				"call1": {
					"id": "call1",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/test1"
				},
				"call2": {
					"id": "call2",
					"method": "HTTP_METHOD_POST",
					"endpoint_path": "/test2"
				}
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	assert.NoError(t, err)
	assert.Len(t, discoveredTools, 2)
	assert.Equal(t, "op_call1", discoveredTools[0].GetName())
	assert.Equal(t, "op_call2", discoveredTools[1].GetName())
}

func TestHTTPUpstream_Register_MissingToolName(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "test-service-missing-tool-name",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{
				"call_id": "test-op-call"
			}],
			"calls": {
				"test-op-call": {
					"id": "test-op-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/test"
				}
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	assert.NoError(t, err)
	assert.Len(t, discoveredTools, 1)
	assert.Equal(t, "op_test-op-call", discoveredTools[0].GetName())
	assert.Len(t, tm.ListTools(), 1)
}

func TestHTTPUpstream_Register(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewManager(nil)
		upstream := NewUpstream(pm)

		configJSON := `{
			"name": "test-service",
			"http_service": {
				"address": "http://127.0.0.1",
				"tools": [{
					"name": "test-op",
					"call_id": "test-op-call"
				}],
				"calls": {
					"test-op-call": {
						"id": "test-op-call",
						"method": "HTTP_METHOD_GET",
						"endpoint_path": "/test"
					}
				}
			}
		}`
		serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		serviceID, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.NoError(t, err)
		expectedKey, _ := util.SanitizeServiceName("test-service")
		assert.Equal(t, expectedKey, serviceID)
		assert.Len(t, discoveredTools, 1)
		p, ok := pool.Get[*client.HTTPClientWrapper](pm, serviceID)
		assert.True(t, ok)
		assert.NotNil(t, p)
	})

	t.Run("nil http service config", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewManager(nil)
		upstream := NewUpstream(pm)

		configJSON := `{"name": "test-service", "grpc_service": {}}`
		serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "http service config is nil")
	})

	t.Run("invalid service name", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewManager(nil)
		upstream := NewUpstream(pm)

		configJSON := `{"name": "", "http_service": {}}`
		serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "id cannot be empty")
	})

	t.Run("invalid scheme", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewManager(nil)
		upstream := NewUpstream(pm)

		configJSON := `{
			"name": "test-service-invalid-scheme",
			"http_service": {
				"address": "file:///etc/passwd",
				"tools": [{
					"name": "test-op",
					"call_id": "test-op-call"
				}],
				"calls": {
					"test-op-call": {
						"id": "test-op-call",
						"method": "HTTP_METHOD_GET",
						"endpoint_path": "/test"
					}
				}
			}
		}`
		serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid http service address scheme")
		assert.Contains(t, err.Error(), "file")
	})

	t.Run("tool registration with fallback operation ID", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewManager(nil)
		upstream := NewUpstream(pm)

		configJSON := `{
			"name": "test-service-fallback",
			"http_service": {
				"address": "http://127.0.0.1",
				"tools": [
					{
						"description": "A test operation",
						"call_id": "test-op-1"
					},
					{
						"description": "",
						"call_id": "test-op-2"
					}
				],
				"calls": {
					"test-op-1": {
						"id": "test-op-1",
						"method": "HTTP_METHOD_GET",
						"endpoint_path": "/test"
					},
					"test-op-2": {
						"id": "test-op-2",
						"method": "HTTP_METHOD_POST",
						"endpoint_path": "/test2"
					}
				}
			}
		}`
		serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
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
		sanitizedName2, _ := util.SanitizeToolName("op_test-op-2")
		toolID2 := serviceID + "." + sanitizedName2
		_, ok = tm.GetTool(toolID2)
		assert.True(t, ok, "Tool with default fallback name should be found, expected %s", toolID2)
	})

	t.Run("authenticator creation fails", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewManager(nil)
		upstream := NewUpstream(pm)

		configJSON := `{
			"name": "auth-fail-service",
			"http_service": {
				"address": "http://127.0.0.1",
				"tools": [{
					"name": "test-op",
					"call_id": "test-op-call"
				}],
				"calls": {
					"test-op-call": {
						"id": "test-op-call",
						"method": "HTTP_METHOD_GET"
					}
				}
			},
			"upstream_auth": {
				"api_key": {
					"value": {"plain_text": ""}
				}
			}
		}`
		serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		require.NoError(t, err)
		assert.Len(t, discoveredTools, 1, "Tool should still be registered even if auth is nil")
	})

	t.Run("registration with connection pool config", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewManager(nil)
		upstream := NewUpstream(pm)

		configJSON := `{
			"name": "test-service-with-pool",
			"http_service": {
				"address": "http://127.0.0.1",
				"tools": [{
					"name": "test-op",
					"call_id": "test-op-call"
				}],
				"calls": {
					"test-op-call": {
						"id": "test-op-call",
						"method": "HTTP_METHOD_GET",
						"endpoint_path": "/test"
					}
				}
			},
			"connection_pool": {
				"max_connections": 50,
				"max_idle_connections": 10
			}
		}`
		serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		// We need to replace the NewHTTPPool function with a mock to check the parameters.
		originalNewHTTPPool := NewHTTPPool
		defer func() { NewHTTPPool = originalNewHTTPPool }()

		var capturedMinSize, capturedMaxSize int
		var capturedIdleTimeout time.Duration
		NewHTTPPool = func(minSize, maxSize int, idleTimeout time.Duration, config *configv1.UpstreamServiceConfig) (pool.Pool[*client.HTTPClientWrapper], error) {
			capturedMinSize = minSize
			capturedMaxSize = maxSize
			capturedIdleTimeout = idleTimeout
			return originalNewHTTPPool(minSize, maxSize, idleTimeout, config)
		}

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.NoError(t, err)

		assert.Equal(t, 10, capturedMinSize)
		assert.Equal(t, 50, capturedMaxSize)
		assert.Equal(t, 90*time.Second, capturedIdleTimeout)
	})

	t.Run("registration with default connection pool config", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewManager(nil)
		upstream := NewUpstream(pm)

		configJSON := `{
			"name": "test-service-with-default-pool",
			"http_service": {
				"address": "http://127.0.0.1",
				"tools": [{
					"name": "test-op",
					"call_id": "test-op-call"
				}],
				"calls": {
					"test-op-call": {
						"id": "test-op-call",
						"method": "HTTP_METHOD_GET",
						"endpoint_path": "/test"
					}
				}
			}
		}`
		serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		// We need to replace the NewHTTPPool function with a mock to check the parameters.
		originalNewHTTPPool := NewHTTPPool
		defer func() { NewHTTPPool = originalNewHTTPPool }()

		var capturedMinSize, capturedMaxSize int
		var capturedIdleTimeout time.Duration
		NewHTTPPool = func(minSize, maxSize int, idleTimeout time.Duration, config *configv1.UpstreamServiceConfig) (pool.Pool[*client.HTTPClientWrapper], error) {
			capturedMinSize = minSize
			capturedMaxSize = maxSize
			capturedIdleTimeout = idleTimeout
			return originalNewHTTPPool(minSize, maxSize, idleTimeout, config)
		}

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.NoError(t, err)

		assert.Equal(t, 10, capturedMinSize)
		assert.Equal(t, 100, capturedMaxSize)
		assert.Equal(t, 90*time.Second, capturedIdleTimeout)
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

	upstream := &Upstream{poolManager: pm}

	configJSON := `{
		"name": "add-tool-fail-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{
				"name": "test-op",
				"call_id": "test-op-call"
			}],
			"calls": {
				"test-op-call": {
					"id": "test-op-call",
					"method": "HTTP_METHOD_GET"
				}
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	var promptManager prompt.ManagerInterface
	var resourceManager resource.ManagerInterface

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, mockTm, promptManager, resourceManager, false)

	require.NoError(t, err)
	assert.Empty(t, discoveredTools, "No tools should be discovered if AddTool fails")
	assert.Empty(t, mockTm.ListTools(), "Tool manager should be empty if AddTool fails")
}

func TestHTTPUpstream_Register_WithReload(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	// Initial registration
	configJSON1 := `{
		"name": "reload-test",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "op1", "call_id": "op1-call"}],
			"calls": {
				"op1-call": {
					"id": "op1-call",
					"method": "HTTP_METHOD_GET"
				}
			}
		}
	}`
	serviceConfig1 := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON1), serviceConfig1))

	serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig1, tm, nil, nil, false)
	require.NoError(t, err)
	assert.Len(t, tm.ListTools(), 1)

	// Reload with a different tool
	configJSON2 := `{
		"name": "reload-test",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "op2", "call_id": "op2-call"}],
			"calls": {
				"op2-call": {
					"id": "op2-call",
					"method": "HTTP_METHOD_GET"
				}
			}
		}
	}`
	serviceConfig2 := configv1.UpstreamServiceConfig_builder{}.Build()
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
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "test-service-invalid-method",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "test-op", "call_id": "test-op-call"}],
			"calls": {
				"test-op-call": {
					"id": "test-op-call",
					"method": 999,
					"endpoint_path": "/test"
				}
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	assert.NoError(t, err)
	assert.Len(t, discoveredTools, 0, "No tools should be registered for an invalid method")
}

// mockToolManager to simulate errors
type mockToolManager struct {
	tool.ManagerInterface
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

func (m *mockToolManager) ListServices() []*tool.ServiceInfo {
	return nil
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

func (m *mockToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}
func (m *mockToolManager) GetServiceInfo(_ string) (*tool.ServiceInfo, bool) {
	return nil, false
}
func (m *mockToolManager) SetMCPServer(_ tool.MCPServerProvider) {}
func (m *mockToolManager) CallTool(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return nil, errors.New("not implemented")
}

func (m *mockToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}

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
			address:      "http://127.0.0.1:8080/",
			endpointPath: "api/v1/test",
			expectedFqn:  "GET http://127.0.0.1:8080/api/v1/test",
		},
		{
			name:         "leading slash in endpoint",
			address:      "http://127.0.0.1:8080",
			endpointPath: "/api/v1/test",
			expectedFqn:  "GET http://127.0.0.1:8080/api/v1/test",
		},
		{
			name:         "both slashes present",
			address:      "http://127.0.0.1:8080/",
			endpointPath: "/api/v1/test",
			expectedFqn:  "GET http://127.0.0.1:8080/api/v1/test",
		},
		{
			name:         "no slashes",
			address:      "http://127.0.0.1:8080",
			endpointPath: "api/v1/test",
			expectedFqn:  "GET http://127.0.0.1:8080/api/v1/test",
		},
		{
			name:         "address with scheme",
			address:      "https://api.example.com",
			endpointPath: "/users",
			expectedFqn:  "GET https://api.example.com/users",
		},
		{
			name:         "double slash bug",
			address:      "http://127.0.0.1:8080/",
			endpointPath: "/api/v1/test",
			expectedFqn:  "GET http://127.0.0.1:8080/api/v1/test",
		},
		{
			name:         "address with path and trailing slash",
			address:      "http://127.0.0.1:8080/base/",
			endpointPath: "api/v1/test",
			expectedFqn:  "GET http://127.0.0.1:8080/base/api/v1/test",
		},
		{
			name:         "trailing slash in endpoint is preserved",
			address:      "http://127.0.0.1:8080",
			endpointPath: "api/v1/test/",
			expectedFqn:  "GET http://127.0.0.1:8080/api/v1/test/",
		},
		{
			name:         "endpoint with query params",
			address:      "http://127.0.0.1:8080",
			endpointPath: "api/v1/test?query=param",
			expectedFqn:  "GET http://127.0.0.1:8080/api/v1/test?query=param",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pm := pool.NewManager()
			tm := tool.NewManager(nil)
			upstream := NewUpstream(pm)

			configJSON := `{
				"name": "url-test-service",
				"http_service": {
					"address": "` + tc.address + `",
					"tools": [{"name": "test-op", "call_id": "test-op-call"}],
					"calls": {
						"test-op-call": {
							"id": "test-op-call",
							"method": "HTTP_METHOD_GET",
							"endpoint_path": "` + tc.endpointPath + `"
						}
					}
				}
			}`
			serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
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

func TestHTTPUpstream_Register_Blocked(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "test-blocked",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [
				{"name": "allowed", "call_id": "c1"},
				{"name": "blocked", "call_id": "c2"}
			],
			"calls": {
				"c1": {"id": "c1", "method": "HTTP_METHOD_GET"},
				"c2": {"id": "c2", "method": "HTTP_METHOD_GET"}
			}
		},
		"call_policies": [
			{
				"rules": [
					{"name_regex": "^blocked$", "action": "DENY"}
				],
				"default_action": "ALLOW"
			}
		]
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)
	assert.Len(t, discoveredTools, 1)
	assert.Equal(t, "allowed", discoveredTools[0].GetName())
}

func TestHTTPUpstream_Shutdown(t *testing.T) {
	pm := pool.NewManager()
	upstream := NewUpstream(pm)
	// Just verify it doesn't panic
	err := upstream.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestHTTPUpstream_CheckHealth(t *testing.T) {
	pm := pool.NewManager()
	u := NewUpstream(pm)

	// Cast to concrete type to access CheckHealth, or use interface if exported
	checker, ok := u.(interface{ CheckHealth(context.Context) error })
	require.True(t, ok)

	// Test with no address (should fail)
	err := checker.CheckHealth(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no address configured")

	// Register with bad address
	configJSON := `{"name": "health-test", "http_service": {"address": "http://localhost:54321"}}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))
	tm := tool.NewManager(nil)

	// Register populates address
	_, _, _, err = u.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	assert.NoError(t, err)

	// Check health (should fail connection)
	err = checker.CheckHealth(context.Background())
	assert.Error(t, err)
	// Error message depends on OS/network stack, but usually contains "connection refused" or "dial tcp"
	assert.True(t, err != nil)
}
