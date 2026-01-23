// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestHTTPUpstream_CheckHealth_ManualAddress(t *testing.T) {
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")

	// This test sets private fields directly to test the fallback path in CheckHealth
	// which is normally unreachable because Register sets the checker.

	// Create a dummy server to check connection against
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	pm := pool.NewManager()
	u := &Upstream{
		poolManager: pm,
		address:     server.URL,
		// checker is intentionally nil
	}

	err := u.CheckHealth(context.Background())
	assert.NoError(t, err)

	// Test with unreachable address
	uBad := &Upstream{
		poolManager: pm,
		address:     "http://localhost:59999", // Likely unreachable
	}
	err = uBad.CheckHealth(context.Background())
	assert.Error(t, err)
}

func TestHTTPUpstream_Register_ExportPolicy_Hidden(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "hidden-tool-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [
				{"name": "visible", "call_id": "c1"},
				{"name": "hidden", "call_id": "c2"}
			],
			"calls": {
				"c1": {"id": "c1", "method": "HTTP_METHOD_GET"},
				"c2": {"id": "c2", "method": "HTTP_METHOD_GET"}
			}
		},
		"tool_export_policy": {
			"rules": [
				{"name_regex": "^visible$", "action": "EXPORT"}
			],
			"default_action": "UNEXPORT"
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	// hidden tool should be skipped
	assert.Len(t, discoveredTools, 1)
	assert.Equal(t, "visible", discoveredTools[0].GetName())
}

func TestHTTPUpstream_Register_BadEndpointPath(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	// A path that fails url.Parse? Hard to find one that isn't empty.
	// Control characters might do it.
	configJSON := `{
		"name": "bad-path-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "bad", "call_id": "bad"}],
			"calls": {
				"bad": {
					"id": "bad",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/\u007f"
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	// Should skip the tool due to parsing error
	assert.Len(t, discoveredTools, 0)
}

func TestHTTPUpstream_Register_PromptAndResourceExportPolicy(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "policy-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"prompts": [
				{"name": "p-visible", "description": "d"},
				{"name": "p-hidden", "description": "d"}
			],
			"resources": [
				{"name": "r-visible", "uri": "file:///v"},
				{"name": "r-hidden", "uri": "file:///h"}
			]
		},
		"prompt_export_policy": {
			"rules": [
				{"name_regex": "^p-visible$", "action": "EXPORT"}
			],
			"default_action": "UNEXPORT"
		},
		"resource_export_policy": {
			"rules": [
				{"name_regex": "^r-visible$", "action": "EXPORT"}
			],
			"default_action": "UNEXPORT"
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
	require.NoError(t, err)

	// Check prompts
	prompts := promptManager.ListPrompts()
	assert.Len(t, prompts, 1)
	// Prompts are namespaced with service ID
	assert.Equal(t, "policy-service.p-visible", prompts[0].Prompt().Name)

	resources := resourceManager.ListResources()
	assert.Len(t, resources, 1)
	assert.Equal(t, "r-visible", resources[0].Resource().Name)
}

// Mock tool manager that fails to add tool
type failToolManager struct {
	*mockToolManager
}

func (m *failToolManager) AddTool(t tool.Tool) error {
	return errors.New("simulated add tool error")
}

func TestHTTPUpstream_Register_AddToolError(t *testing.T) {
	pm := pool.NewManager()
	tm := &failToolManager{mockToolManager: newMockToolManager()}
	resourceManager := resource.NewManager()
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "add-fail-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "t1", "call_id": "c1"}],
			"calls": {
				"c1": {"id": "c1", "method": "HTTP_METHOD_GET"}
			},
            "resources": [
                {
                    "name": "res",
                    "uri": "http://res",
                    "dynamic": {
                        "http_call": {"id": "c1"}
                    }
                }
            ]
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, resourceManager, false)
	require.NoError(t, err)

	// Tool addition failed, so it should not be in discoveredTools
	assert.Len(t, discoveredTools, 0)
    // Resource addition should fail too because tool is missing
    assert.Len(t, resourceManager.ListResources(), 0)
}

func TestHTTPUpstream_Register_DynamicResource_InvalidCall(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "dynamic-res-fail",
		"http_service": {
			"address": "http://127.0.0.1",
			"resources": [
				{
					"name": "res",
					"uri": "http://res",
					"dynamic": {
						"http_call": {"id": "non-existent"}
					}
				}
			]
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)
    // Should log error but succeed registration
}

func TestHTTPUpstream_URLConstruction_DoubleSlashWithHost(t *testing.T) {
	testCases := []struct {
		name          string
		address       string
		endpointPath  string
		expectedFqn   string
	}{
		{
			name:         "endpoint starting with double slash interpreted as host",
			address:      "http://example.com/api",
			endpointPath: "//foo/bar",
			// url.Parse("//foo/bar") -> Host=foo, Path=/bar.
			// Our fix should reconstruct it as Path //foo/bar and append to base.
			// Base: http://example.com/api (no trailing slash)
			// Resolved: http://example.com/api//foo/bar
			expectedFqn:  "GET http://example.com/api//foo/bar",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pm := pool.NewManager()
			tm := tool.NewManager(nil)
			upstream := NewUpstream(pm)

			configJSON := `{
				"name": "double-slash-host-service",
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
			serviceConfig := &configv1.UpstreamServiceConfig{}
			require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

			serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
			assert.NoError(t, err)

			sanitizedToolName, _ := util.SanitizeToolName("test-op")
			toolID := serviceID + "." + sanitizedToolName
			registeredTool, ok := tm.GetTool(toolID)
			assert.True(t, ok)
			assert.NotNil(t, registeredTool)

			actualFqn := registeredTool.Tool().GetUnderlyingMethodFqn()
			assert.Equal(t, tc.expectedFqn, actualFqn)
		})
	}
}

func TestHTTPUpstream_CheckHealth_WithChecker(t *testing.T) {
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")

	// Create a dummy server that returns 200 OK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "health-check-service",
		"http_service": {
			"address": "` + server.URL + `",
			"health_check": {
				"url": "` + server.URL + `/health",
				"expected_code": 200
			},
			"tools": [{"name": "t1", "call_id": "c1"}],
			"calls": {
				"c1": {"id": "c1", "method": "HTTP_METHOD_GET"}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	// Now check health. u.checker should be non-nil.
	// We need to cast to concrete type or interface to check CheckHealth
	checker, ok := upstream.(interface{ CheckHealth(context.Context) error })
	require.True(t, ok)

	err = checker.CheckHealth(context.Background())
	assert.NoError(t, err)
}

func TestHTTPUpstream_CheckHealth_WithChecker_Fail(t *testing.T) {
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")

	// Create a dummy server that returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "health-check-fail-service",
		"http_service": {
			"address": "` + server.URL + `",
			"health_check": {
				"url": "` + server.URL + `/health",
				"expected_code": 200
			},
			"tools": [{"name": "t1", "call_id": "c1"}],
			"calls": {
				"c1": {"id": "c1", "method": "HTTP_METHOD_GET"}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	checker, ok := upstream.(interface{ CheckHealth(context.Context) error })
	require.True(t, ok)

	err = checker.CheckHealth(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "health check failed")
}

func TestHTTPUpstream_Register_InvalidCallPolicy(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "invalid-policy-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "t", "call_id": "c"}],
			"calls": {
				"c": {"id": "c", "method": "HTTP_METHOD_GET"}
			}
		},
		"call_policies": [
			{
				"rules": [
					{"name_regex": "(", "action": "ALLOW"}
				],
				"default_action": "DENY"
			}
		]
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	// Register logs error and returns nil tools, but no error
	assert.NoError(t, err)
	assert.Nil(t, discoveredTools)
}
