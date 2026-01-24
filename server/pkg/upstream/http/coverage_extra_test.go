package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestCheckHealth_WithChecker(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	u := NewUpstream(pm)

	// Config with health check INSIDE http_service
	configJSON := `{
		"name": "health-check-service",
		"http_service": {
			"address": "http://127.0.0.1:54321",
			"health_check": {
				"url": "http://127.0.0.1:54321/health"
			}
		}
	}`
	// Address is likely unreachable, so check should fail.
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	u.Register(context.Background(), serviceConfig, tm, nil, nil, false)

	hc, ok := u.(upstream.HealthChecker)
	require.True(t, ok)

	// This should exercise the u.checker != nil path
	err := hc.CheckHealth(context.Background())
	// We expect error because 127.0.0.1:54321 is likely down
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "health check failed")
}

func TestCheckHealth_Success(t *testing.T) {
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	u := NewUpstream(pm)

	configJSON := `{
		"name": "health-check-success",
		"http_service": {
			"address": "` + server.URL + `",
			"health_check": {
				"url": "` + server.URL + `",
				"expected_code": 200
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	u.Register(context.Background(), serviceConfig, tm, nil, nil, false)

	hc, ok := u.(upstream.HealthChecker)
	require.True(t, ok)

	err := hc.CheckHealth(context.Background())
	assert.NoError(t, err)
}

func TestRegister_CoverageEdges(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	u := NewUpstream(pm)

	configJSON := `{
		"name": "coverage-service",
		"http_service": {
			"address": "http://example.com",
			"tools": [
				{
					"name": "disabled-tool",
					"call_id": "disabled-call",
					"disable": true,
					"description": "Disabled tool"
				},
				{
					"name": "hidden-tool",
					"call_id": "hidden-call",
					"description": "Hidden tool"
				},
				{
					"name": "normal-tool",
					"call_id": "normal-call",
					"description": "Normal tool"
				}
			],
			"calls": {
				"disabled-call": {
					"id": "disabled-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/disabled"
				},
				"hidden-call": {
					"id": "hidden-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/hidden"
				},
				"normal-call": {
					"id": "normal-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/normal"
				},
				"missing-def-call": {
					"id": "missing-def-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/missing"
				}
			}
		},
		"tool_export_policy": {
			"rules": [
				{"name_regex": "^hidden-tool$", "action": "UNEXPORT"}
			]
		}
	}`

	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	serviceID, tools, _, err := u.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	assert.NoError(t, err)
	assert.NotEmpty(t, serviceID)

	assert.Len(t, tools, 1)
	if len(tools) > 0 {
		assert.Equal(t, "normal-tool", tools[0].GetName())
	}

	_, ok := tm.GetTool(serviceID + ".disabled-tool")
	assert.False(t, ok, "disabled tool should not be registered")

	_, ok = tm.GetTool(serviceID + ".hidden-tool")
	assert.False(t, ok, "hidden tool should not be registered")
}

func TestRegister_CallPolicyError(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	u := NewUpstream(pm)

	configJSON := `{
		"name": "policy-error-service",
		"http_service": {
			"address": "http://example.com",
			"tools": [
				{
					"name": "tool1",
					"call_id": "call1",
					"description": "Tool 1"
				}
			],
			"calls": {
				"call1": {
					"id": "call1",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/path1"
				}
			}
		},
		"call_policies": [
			{
				"rules": [
					{"name_regex": "(", "action": "ALLOW"}
				]
			}
		]
	}`

	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	serviceID, tools, _, err := u.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	assert.NoError(t, err)
	assert.NotEmpty(t, serviceID)
	assert.Empty(t, tools)
}

func TestRegister_EndpointParseError(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	u := NewUpstream(pm)

	configJSON := `{
		"name": "endpoint-error-service",
		"http_service": {
			"address": "http://example.com",
			"tools": [
				{
					"name": "tool1",
					"call_id": "call1"
				}
			],
			"calls": {
				"call1": {
					"id": "call1",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": ":invalid-url-path"
				}
			}
		}
	}`

	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	// Inject invalid path
	serviceConfig.GetHttpService().Calls["call1"].EndpointPath = proto.String(string(rune(0x7f)))

	_, tools, _, err := u.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	assert.NoError(t, err)
	// Tool creation should be skipped for this call
	assert.Empty(t, tools)
}

func TestRegister_DynamicResource_ToolNotFound(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	rm := resource.NewManager()
	u := NewUpstream(pm)

	configJSON := `{
		"name": "dynamic-res-error-service",
		"http_service": {
			"address": "http://example.com",
			"tools": [
				{
					"name": "tool1",
					"call_id": "call1"
				}
			],
			"calls": {
				"call1": {
					"id": "call1",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/path1"
				},
				"call2": {
					"id": "call2",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/path2"
				}
			},
			"resources": [
				{
					"name": "res1",
					"uri": "http://res1",
					"dynamic": {
						"http_call": {
							"id": "call2"
						}
					}
				}
			]
		}
	}`
	// call2 exists in Calls, but NOT in Tools. So no tool is registered for it.
	// Resource refers to call2. It should fail to find the tool.

	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := u.Register(context.Background(), serviceConfig, tm, nil, rm, false)
	assert.NoError(t, err)

	// Resource should not be registered
	_, ok := rm.GetResource("http://res1")
	assert.False(t, ok)
}

func TestRegister_InputSchemaMerging(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	u := NewUpstream(pm)

	configJSON := `{
		"name": "schema-merge-service",
		"http_service": {
			"address": "http://example.com",
			"tools": [
				{
					"name": "tool1",
					"call_id": "call1"
				}
			],
			"calls": {
				"call1": {
					"id": "call1",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/path1",
					"parameters": [
						{
							"schema": {
								"name": "p1",
								"type": "STRING",
								"is_required": true
							}
						}
					],
					"input_schema": {
						"type": "object",
						"properties": {
							"p2": { "type": "integer" }
						},
						"required": ["p2"]
					}
				}
			}
		}
	}`

	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	serviceID, _, _, err := u.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	assert.NoError(t, err)

	sanitizedToolName, _ := util.SanitizeToolName("tool1")
	toolID := serviceID + "." + sanitizedToolName
	registeredTool, ok := tm.GetTool(toolID)
	assert.True(t, ok)

	schema := registeredTool.Tool().GetAnnotations().GetInputSchema()
	require.NotNil(t, schema)

	// Check properties: p1 (from parameters) and p2 (from input_schema) should be present
	props := schema.Fields["properties"].GetStructValue().Fields
	assert.Contains(t, props, "p1")
	assert.Contains(t, props, "p2")

	// Check required: p1 and p2 should be present
	required := schema.Fields["required"].GetListValue().Values
	var reqStrs []string
	for _, v := range required {
		reqStrs = append(reqStrs, v.GetStringValue())
	}
	assert.Contains(t, reqStrs, "p1")
	assert.Contains(t, reqStrs, "p2")
}

func TestRegister_InputSchema_NoParameters(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	u := NewUpstream(pm)

	configJSON := `{
		"name": "schema-no-param-service",
		"http_service": {
			"address": "http://example.com",
			"tools": [
				{
					"name": "tool1",
					"call_id": "call1"
				}
			],
			"calls": {
				"call1": {
					"id": "call1",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/path1",
					"input_schema": {
						"type": "object",
						"properties": {
							"p2": { "type": "integer" }
						}
					}
				}
			}
		}
	}`

	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	serviceID, _, _, err := u.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	assert.NoError(t, err)

	sanitizedToolName, _ := util.SanitizeToolName("tool1")
	toolID := serviceID + "." + sanitizedToolName
	registeredTool, ok := tm.GetTool(toolID)
	assert.True(t, ok)

	schema := registeredTool.Tool().GetAnnotations().GetInputSchema()
	require.NotNil(t, schema)
	// Just verify p2 exists
	props := schema.Fields["properties"].GetStructValue().Fields
	assert.Contains(t, props, "p2")
}

func TestRegister_InputSchema_NoRequired(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	u := NewUpstream(pm)

	configJSON := `{
		"name": "schema-no-req-service",
		"http_service": {
			"address": "http://example.com",
			"tools": [
				{
					"name": "tool1",
					"call_id": "call1"
				}
			],
			"calls": {
				"call1": {
					"id": "call1",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/path1",
					"parameters": [
						{
							"schema": {
								"name": "p1",
								"type": "STRING",
								"is_required": true
							}
						}
					],
					"input_schema": {
						"type": "object",
						"properties": {
							"p2": { "type": "integer" }
						}
					}
				}
			}
		}
	}`

	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	serviceID, _, _, err := u.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	assert.NoError(t, err)

	sanitizedToolName, _ := util.SanitizeToolName("tool1")
	toolID := serviceID + "." + sanitizedToolName
	registeredTool, ok := tm.GetTool(toolID)
	assert.True(t, ok)

	schema := registeredTool.Tool().GetAnnotations().GetInputSchema()
	require.NotNil(t, schema)

	// Check required: p1 should be present (added from parameters)
	required := schema.Fields["required"].GetListValue().Values
	var reqStrs []string
	for _, v := range required {
		reqStrs = append(reqStrs, v.GetStringValue())
	}
	assert.Contains(t, reqStrs, "p1")
}

func TestRegister_InvalidQueryKey(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	u := NewUpstream(pm)

	configJSON := `{
		"name": "invalid-query-key-service",
		"http_service": {
			"address": "http://example.com/api?a=1",
			"tools": [{"name": "test-op", "call_id": "test-op-call"}],
			"calls": {
				"test-op-call": {
					"id": "test-op-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/test?%GG=value"
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	serviceID, _, _, err := u.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	assert.NoError(t, err)

	sanitizedToolName, _ := util.SanitizeToolName("test-op")
	toolID := serviceID + "." + sanitizedToolName
	registeredTool, ok := tm.GetTool(toolID)
	assert.True(t, ok)

	actualFqn := registeredTool.Tool().GetUnderlyingMethodFqn()
	// %GG=value has invalid key. Should be treated as raw and appended.
	// Base "a=1" should be preserved.
	// Result: a=1&%GG=value
	assert.Equal(t, "GET http://example.com/api/test?a=1&%GG=value", actualFqn)
}

func TestHTTPUpstream_QueryMerge_InvalidOverride(t *testing.T) {
	testCases := []struct {
		name         string
		address      string
		endpointPath string
		expectedFqn  string
	}{
		{
			name:         "endpoint with invalid encoding should override base param",
			address:      "http://example.com/api?q=default",
			endpointPath: "/v1/test?q=%GG",
			// The bug is that we expect "q=%GG" (override), but we get "q=default&q=%GG" (append)
			expectedFqn: "GET http://example.com/api/v1/test?q=%GG",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pm := pool.NewManager()
			tm := tool.NewManager(nil)
			upstream := NewUpstream(pm)

			configJSON := `{
				"name": "query-invalid-override-service",
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
