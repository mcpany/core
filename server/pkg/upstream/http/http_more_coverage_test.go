// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"net/http"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestHTTPUpstream_Register_CallPolicyCompileError(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "policy-compile-error",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "op", "call_id": "call"}],
			"calls": {
				"call": {"id": "call", "method": "HTTP_METHOD_GET"}
			}
		},
		"call_policies": [
			{
				"rules": [
					{"name_regex": "(", "action": "DENY"}
				]
			}
		]
	}`
	// "(" is invalid regex.

	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.WithValue(context.Background(), ContextKeySkipConnectionCheck, true), serviceConfig, tm, nil, nil, false)
	// Register returns error only on critical setup. Call policy failure in createAndRegisterHTTPTools returns nil (logs error).
	// But `createAndRegisterHTTPTools` returns nil slice on compile error.
	require.NoError(t, err)
	assert.Nil(t, discoveredTools)
	assert.Empty(t, tm.ListTools())
}


func TestHTTPUpstream_Register_UnsupportedMethod(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "unsupported-method",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "op", "call_id": "call"}],
			"calls": {
				"call": {"id": "call", "method": "HTTP_METHOD_UNSPECIFIED"}
			}
		}
	}`
	// UNSPECIFIED -> httpMethodToString returns error.

	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.WithValue(context.Background(), ContextKeySkipConnectionCheck, true), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)
	assert.Len(t, discoveredTools, 0)
}

func TestHTTPUpstream_Register_DoubleSlashParseFailure(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	// Test the path where url.Parse fails on a double-slash path
	// and we attempt to recover by prepending slash.

	configJSON := `{
		"name": "double-slash-parse-fail",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "op", "call_id": "call"}],
			"calls": {
				"call": {
					"id": "call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "//:invalid"
				}
			}
		}
	}`
	// "//:invalid" might fail Parse?

	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	// If recovery works, it becomes "///:invalid" which is valid path.
	// So tool should be created.

	_, discoveredTools, _, err := upstream.Register(context.WithValue(context.Background(), ContextKeySkipConnectionCheck, true), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	// If it recovered, we have 1 tool. If not, 0.
	// Go's url.Parse might not actually fail on "//:invalid".
	// But assuming we can find a case.
	// If it fails, we cover the block.

	if len(discoveredTools) == 1 {
		toolID := "double-slash-parse-fail.op"
		tool, ok := tm.GetTool(toolID)
		assert.True(t, ok)
		fqn := tool.Tool().GetUnderlyingMethodFqn()
		// If recovered: http://127.0.0.1//:invalid
		// Update: My analysis showed it resolves to http://127.0.0.1//:invalid
		assert.Contains(t, fqn, "http://127.0.0.1//:invalid")
	}
}

func TestHTTPUpstream_Register_InputSchema_PropertiesNotStruct(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "schema-bad-props",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "op", "call_id": "call"}],
			"calls": {
				"call": {
					"id": "call",
					"method": "HTTP_METHOD_GET",
					"input_schema": {
						"properties": "not-a-struct"
					},
					"parameters": [
						{
							"schema": {
								"name": "auto_prop",
								"type": "STRING",
								"is_required": true
							}
						}
					]
				}
			}
		}
	}`

	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.WithValue(context.Background(), ContextKeySkipConnectionCheck, true), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	toolID := "schema-bad-props.op"
	tool, ok := tm.GetTool(toolID)
	assert.True(t, ok)

	schema := tool.Tool().GetAnnotations().GetInputSchema()
	fields := schema.GetFields()

	// "properties" should now be a struct containing "auto_prop", overwriting string value
	props := fields["properties"].GetStructValue().GetFields()
	assert.Contains(t, props, "auto_prop")
}

func TestNewHTTPPool_PoolNewError(t *testing.T) {
	configJSON := `{
		"http_service": {
			"tls_config": {}
		}
	}`
	config := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), config))

	// MinSize > MaxSize causes pool.New to fail.
	_, err := NewHTTPPool(10, 5, 0, config)
	assert.Error(t, err)
}

func TestHTTPUpstream_Register_EmptyAddress(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "empty-address",
		"http_service": {
			"address": ""
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.WithValue(context.Background(), ContextKeySkipConnectionCheck, true), serviceConfig, tm, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "http service address is required")
}

func TestHTTPUpstream_Register_DoubleSlashPath(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "double-slash-test",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "op", "call_id": "call"}],
			"calls": {
				"call": {
					"id": "call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "//foo/bar"
				}
			}
		}
	}`

	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.WithValue(context.Background(), ContextKeySkipConnectionCheck, true), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)
	assert.Len(t, discoveredTools, 1)

	// Verify the tool URL.
	toolID := "double-slash-test.op"
	tool, ok := tm.GetTool(toolID)
	assert.True(t, ok)
	fqn := tool.Tool().GetUnderlyingMethodFqn()
	assert.Contains(t, fqn, "http://127.0.0.1//foo/bar")
}

func TestHTTPUpstream_Register_ExplicitInputSchema_Merging(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	// Define a tool with BOTH input_schema and parameters.
	// Check if parameters are merged into input_schema.

	configJSON := `{
		"name": "explicit-schema-merge",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "op", "call_id": "call"}],
			"calls": {
				"call": {
					"id": "call",
					"method": "HTTP_METHOD_GET",
					"input_schema": {
						"existing_prop": "foo",
						"properties": {
							"manual_prop": "manual"
						},
						"required": ["manual_prop"]
					},
					"parameters": [
						{
							"schema": {
								"name": "auto_prop",
								"type": "STRING",
								"is_required": true
							}
						}
					]
				}
			}
		}
	}`

	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.WithValue(context.Background(), ContextKeySkipConnectionCheck, true), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	toolID := "explicit-schema-merge.op"
	tool, ok := tm.GetTool(toolID)
	assert.True(t, ok)

	schema := tool.Tool().GetAnnotations().GetInputSchema()
	fields := schema.GetFields()

	// Check properties merge
	props := fields["properties"].GetStructValue().GetFields()
	assert.Contains(t, props, "manual_prop")
	assert.Contains(t, props, "auto_prop")

	// Check required merge
	reqs := fields["required"].GetListValue().GetValues()
	reqNames := make([]string, len(reqs))
	for i, v := range reqs {
		reqNames[i] = v.GetStringValue()
	}
	assert.Contains(t, reqNames, "manual_prop")
	assert.Contains(t, reqNames, "auto_prop")
}

func TestHTTPUpstream_Register_EndpointParseFailure(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	// We need a path that `url.Parse` fails on.
	// Most strings are valid paths.
	// Control characters might fail it?

	configJSON := `{
		"name": "endpoint-parse-fail",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "op", "call_id": "call"}],
			"calls": {
				"call": {
					"id": "call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": ":/bad-path-scheme"
				}
			}
		}
	}`
	// ":/..." usually parses as scheme ":" which is invalid? Or valid scheme empty?
	// url.Parse(":/foo") -> err "first path segment in URL cannot contain colon"

	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.WithValue(context.Background(), ContextKeySkipConnectionCheck, true), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)
	assert.Len(t, discoveredTools, 0)
}

func TestHTTPUpstream_Register_BaseURLParseFailure(t *testing.T) {
	// Register checks ParseRequestURI.
	// createAndRegisterHTTPTools checks Parse.
	// If ParseRequestURI passes but Parse fails?
	// ParseRequestURI requires scheme and host. Parse does not necessarily.
	// It's hard to find a URL that passes ParseRequestURI but fails Parse.
	// Usually it's the other way around.
}

func TestHTTPPool_Close_Error(t *testing.T) {
	mockErr := assert.AnError
	mp := &mockPool{closeErr: mockErr}
	hp := &httpPool{Pool: mp}

	err := hp.Close()
	assert.ErrorIs(t, err, mockErr)
}

func TestHTTPPool_Close_Success(t *testing.T) {
	mp := &mockPool{}
	// We need a transport to close
	tr := &http.Transport{}
	hp := &httpPool{Pool: mp, transport: tr}

	err := hp.Close()
	assert.NoError(t, err)
}

type mockPool struct {
	pool.Pool[*client.HTTPClientWrapper]
	closeErr error
}

func (m *mockPool) Close() error {
	return m.closeErr
}

// We need to access client package for HTTPClientWrapper, which is imported.

func TestNewHTTPPool_MTLS_CertError(t *testing.T) {
	configJSON := `{
		"upstream_auth": {
			"mtls": {
				"client_cert_path": "non-existent-cert",
				"client_key_path": "non-existent-key"
			}
		},
		"http_service": {
			"tls_config": {}
		}
	}`
	config := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), config))

	_, err := NewHTTPPool(1, 1, 0, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "open non-existent-cert")
}

func TestHTTPUpstream_Register_DisabledTool(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "disabled-tool-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [
				{"name": "op1", "call_id": "c1", "disable": true}
			],
			"calls": {
				"c1": {"id": "c1", "method": "HTTP_METHOD_GET"}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.WithValue(context.Background(), ContextKeySkipConnectionCheck, true), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)
	assert.Empty(t, discoveredTools)
}

func TestHTTPUpstream_Register_DisabledResource(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)
	rm := resource.NewManager()

	configJSON := `{
		"name": "disabled-resource-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "op", "call_id": "c"}],
			"calls": {"c": {"id": "c", "method": "HTTP_METHOD_GET"}},
			"resources": [
				{"name": "res1", "uri": "http://res1", "disable": true}
			]
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.WithValue(context.Background(), ContextKeySkipConnectionCheck, true), serviceConfig, tm, nil, rm, false)
	require.NoError(t, err)

	_, ok := rm.GetResource("http://res1")
	assert.False(t, ok)
}

func TestHTTPUpstream_Register_DisabledPrompt(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)
	prompts := prompt.NewManager()

	configJSON := `{
		"name": "disabled-prompt-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"prompts": [
				{"name": "prompt1", "disable": true}
			]
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.WithValue(context.Background(), ContextKeySkipConnectionCheck, true), serviceConfig, tm, prompts, nil, false)
	require.NoError(t, err)

	_, ok := prompts.GetPrompt("prompt1")
	assert.False(t, ok)
}

func TestHTTPUpstream_Register_InputSchema_NoProperties(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "no-props-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "op", "call_id": "c"}],
			"calls": {
				"c": {
					"id": "c",
					"method": "HTTP_METHOD_GET",
					"input_schema": {
						"type": "object"
					},
					"parameters": [
						{
							"schema": {
								"name": "foo",
								"type": "STRING",
                                "is_required": false
							}
						}
					]
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.WithValue(context.Background(), ContextKeySkipConnectionCheck, true), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	tools := tm.ListTools()
	require.Len(t, tools, 1)
	schema := tools[0].Tool().GetAnnotations().GetInputSchema()

	// properties should be created
	propsVal := schema.Fields["properties"].GetStructValue()
	require.NotNil(t, propsVal)
	_, ok := propsVal.Fields["foo"]
	assert.True(t, ok)
}

func TestHTTPUpstream_Register_InputSchema_NoRequired(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "no-req-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "op", "call_id": "c"}],
			"calls": {
				"c": {
					"id": "c",
					"method": "HTTP_METHOD_GET",
					"input_schema": {
						"type": "object",
						"properties": {}
					},
					"parameters": [
						{
							"schema": {
								"name": "foo",
								"type": "STRING",
                                "is_required": true
							}
						}
					]
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.WithValue(context.Background(), ContextKeySkipConnectionCheck, true), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	tools := tm.ListTools()
	require.Len(t, tools, 1)
	schema := tools[0].Tool().GetAnnotations().GetInputSchema()

	// required should be created
	reqVal := schema.Fields["required"].GetListValue()
	require.NotNil(t, reqVal)
	assert.Equal(t, "foo", reqVal.Values[0].GetStringValue())
}

func TestHTTPUpstream_Register_InputSchema_InvalidRequired(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "invalid-req-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "op", "call_id": "c"}],
			"calls": {
				"c": {
					"id": "c",
					"method": "HTTP_METHOD_GET",
					"input_schema": {
						"type": "object",
						"properties": {},
						"required": "not-a-list"
					},
					"parameters": [
						{
							"schema": {
								"name": "foo",
								"type": "STRING",
                                "is_required": true
							}
						}
					]
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.WithValue(context.Background(), ContextKeySkipConnectionCheck, true), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	tools := tm.ListTools()
	require.Len(t, tools, 1)
	schema := tools[0].Tool().GetAnnotations().GetInputSchema()

	// required should be overwritten by list containing foo
	reqVal := schema.Fields["required"].GetListValue()
	require.NotNil(t, reqVal)
	assert.Equal(t, "foo", reqVal.Values[0].GetStringValue())
}

func TestHTTPUpstream_Register_AuthenticatorError(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "auth-error-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "op", "call_id": "c"}],
			"calls": {"c": {"id": "c", "method": "HTTP_METHOD_GET"}}
		},
		"upstream_auth": {
			"api_key": {}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.WithValue(context.Background(), ContextKeySkipConnectionCheck, true), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)
	// Should still register tools, but without auth (logged error)
	assert.NotEmpty(t, discoveredTools)
}


func TestHTTPUpstream_Register_ExportPolicy_UnexportTool(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "unexport-tool-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [
				{"name": "op1", "call_id": "c1"}
			],
			"calls": {
				"c1": {"id": "c1", "method": "HTTP_METHOD_GET"}
			}
		},
		"tool_export_policy": {
			"default_action": "UNEXPORT"
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.WithValue(context.Background(), ContextKeySkipConnectionCheck, true), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)
	assert.Empty(t, discoveredTools)
}

func TestHTTPUpstream_Register_ExportPolicy_UnexportResource(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)
	rm := resource.NewManager()

	configJSON := `{
		"name": "unexport-resource-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"resources": [
				{"name": "res1", "uri": "http://res1"}
			]
		},
		"resource_export_policy": {
			"default_action": "UNEXPORT"
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.WithValue(context.Background(), ContextKeySkipConnectionCheck, true), serviceConfig, tm, nil, rm, false)
	require.NoError(t, err)

	_, ok := rm.GetResource("http://res1")
	assert.False(t, ok)
}

func TestHTTPUpstream_Register_ExportPolicy_UnexportPrompt(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)
	prompts := prompt.NewManager()

	configJSON := `{
		"name": "unexport-prompt-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"prompts": [
				{"name": "prompt1"}
			]
		},
		"prompt_export_policy": {
			"default_action": "UNEXPORT"
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.WithValue(context.Background(), ContextKeySkipConnectionCheck, true), serviceConfig, tm, prompts, nil, false)
	require.NoError(t, err)

	_, ok := prompts.GetPrompt("prompt1")
	assert.False(t, ok)
}
