// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestHTTPUpstream_CheckHealth_Success(t *testing.T) {
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")

	// Start a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	pm := pool.NewManager()
	u := NewUpstream(pm)

	// Register with valid address to populate u.address
	configJSON := `{
		"name": "health-success-test",
		"http_service": {
			"address": "` + server.URL + `",
			"health_check": {
				"url": "` + server.URL + `"
			},
			"tools": [{"name": "test-op", "call_id": "test-op-call"}],
			"calls": {
				"test-op-call": {
					"id": "test-op-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/"
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))
	tm := tool.NewManager(nil)

	// Register
	_, _, _, err := u.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	assert.NoError(t, err)

	// CheckHealth should succeed via checker
	checker, ok := interface{}(u).(interface{ CheckHealth(context.Context) error })
	require.True(t, ok)
	err = checker.CheckHealth(context.Background())
	assert.NoError(t, err)

	// Hack: We want to test the path where u.checker is nil.
	// Register creates u.checker.
	// But `Upstream` struct fields are unexported.
	// We can create a NEW upstream manually and set only what we need?
	// No, we can't set unexported fields from another package (or even same package if not using internal access, but we are in same package).
	// We are in `http` package, so we CAN access unexported fields.

	u2 := &Upstream{
		poolManager: pm,
		address:     server.URL,
		// checker is nil
	}

	checker2, ok := interface{}(u2).(interface{ CheckHealth(context.Context) error })
	require.True(t, ok)

	err = checker2.CheckHealth(context.Background())
	assert.NoError(t, err)
}

func TestHTTPUpstream_CheckHealth_CheckerFailure(t *testing.T) {
	pm := pool.NewManager()
	u := NewUpstream(pm)

	// Register with address that fails health check (e.g. 500)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	configJSON := `{
		"name": "health-fail-test",
		"http_service": {
			"address": "` + server.URL + `",
			"health_check": {
				"url": "` + server.URL + `"
			},
			"tools": [{"name": "test-op", "call_id": "test-op-call"}],
			"calls": {
				"test-op-call": {
					"id": "test-op-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/"
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))
	tm := tool.NewManager(nil)

	// Register
	_, _, _, err := u.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	assert.NoError(t, err)

	checker, ok := interface{}(u).(interface{ CheckHealth(context.Context) error })
	require.True(t, ok)

	// Should fail because server returns 500
	err = checker.CheckHealth(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "health check failed")
}

func TestHTTPUpstream_Register_CompileCallPoliciesError(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "policy-fail-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "test-op", "call_id": "test-op-call"}],
			"calls": {
				"test-op-call": {
					"id": "test-op-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/test"
				}
			}
		},
		"call_policies": [
			{
				"rules": [
					{"name_regex": "(", "action": "DENY"}
				],
				"default_action": "ALLOW"
			}
		]
	}`
	// "(" is an invalid regex.

	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)

	// Register should verify policy compilation?
	// No, Register logs error and returns nil tools if policy compilation fails inside createAndRegisterHTTPTools.
	// BUT, Register calls createAndRegisterHTTPTools which returns (..., nil, nil, err)?
	// Wait, createAndRegisterHTTPTools returns []*ToolDefinition. It does NOT return error.
	// Inside createAndRegisterHTTPTools:
	// compiledCallPolicies, err := tool.CompileCallPolicies(callPolicies)
	// if err != nil { log.Error(...); return nil }

	// So Register succeeds but returns no tools.
	assert.NoError(t, err)
	assert.Empty(t, discoveredTools)
}

func TestHTTPUpstream_Register_ExportPolicy_Unexport(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "export-policy-test",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [
				{"name": "op1", "call_id": "op1-call"}
			],
			"calls": {
				"op1-call": {
					"id": "op1-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/test"
				}
			},
			"prompts": [
				{"name": "prompt1"}
			],
			"resources": [
				{"name": "res1", "uri": "http://res1"}
			]
		},
		"tool_export_policy": {
			"default_action": "UNEXPORT"
		},
		"prompt_export_policy": {
			"default_action": "UNEXPORT"
		},
		"resource_export_policy": {
			"default_action": "UNEXPORT"
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
	require.NoError(t, err)

	// Tools should be empty
	assert.Empty(t, discoveredTools)

	// Prompts should be empty (not registered)
	assert.Empty(t, promptManager.ListPrompts())

	// Resources should be empty
	_, ok := resourceManager.GetResource("http://res1")
	assert.False(t, ok)

	// We should also test that the loop continues.
}

func TestHTTPUpstream_Register_EndpointPathParseError(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	// \x7f is a control character, valid in JSON string but invalid in URL?
	// net/url Parse validates some characters.
	// Let's try to put something that url.Parse rejects.
	// It accepts almost anything in path if not encoded.
	// But `ctl` characters might be rejected.

	configJSON := `{
		"name": "url-parse-fail",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "test-op", "call_id": "test-op-call"}],
			"calls": {
				"test-op-call": {
					"id": "test-op-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/test\u007f"
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	assert.NoError(t, err)
	// Should be empty because parsing endpoint path failed
	assert.Empty(t, discoveredTools)
}

func TestHTTPUpstream_Register_DynamicResource_CallIdNotFound(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "dr-call-fail",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "op1", "call_id": "op1-call"}],
			"calls": {
				"op1-call": {
					"id": "op1-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/test"
				}
			},
			"resources": [
				{
					"name": "res1",
					"uri": "http://res1",
					"dynamic": {
						"http_call": {"id": "non-existent-call"}
					}
				}
			]
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))
	resourceManager := resource.NewManager()

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, resourceManager, false)
	assert.NoError(t, err)

	// Resource should not be registered
	_, ok := resourceManager.GetResource("http://res1")
	assert.False(t, ok)
}

func TestHTTPUpstream_Register_WithIdleTimeout(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "idle-timeout-test",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "test-op", "call_id": "test-op-call"}],
			"calls": {
				"test-op-call": {
					"id": "test-op-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/test"
				}
			},
			"connection_pool": {
				"idle_timeout": "60s"
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	// Mock NewHTTPPool to verify idleTimeout
	originalNewHTTPPool := NewHTTPPool
	defer func() { NewHTTPPool = originalNewHTTPPool }()

	var capturedIdleTimeout time.Duration
	NewHTTPPool = func(minSize, maxSize int, idleTimeout time.Duration, config *configv1.UpstreamServiceConfig) (pool.Pool[*client.HTTPClientWrapper], error) {
		capturedIdleTimeout = idleTimeout
		return originalNewHTTPPool(minSize, maxSize, idleTimeout, config)
	}

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	assert.NoError(t, err)

	assert.Equal(t, 60*time.Second, capturedIdleTimeout)
}

func TestHTTPUpstream_Register_DynamicResource_ToolNotFound(t *testing.T) {
	pm := pool.NewManager()
	mockTm := newMockToolManager()
	// We want AddTool to fail so the tool is not registered.
	// But we also want to test the dynamic resource logic.
	// The code iterates tools first, then resources.
	// If AddTool fails, the tool is not in mockTm.
	// Then resource loop runs. It finds the call ID in config, gets tool name.
	// Then it calls toolManager.GetTool.
	// Since tool wasn't added, GetTool returns false.

	mockTm.addError = assert.AnError // Force AddTool error

	upstream := &Upstream{poolManager: pm}

	configJSON := `{
		"name": "dr-tool-fail",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "op1", "call_id": "op1-call"}],
			"calls": {
				"op1-call": {
					"id": "op1-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/test"
				}
			},
			"resources": [
				{
					"name": "res1",
					"uri": "http://res1",
					"dynamic": {
						"http_call": {"id": "op1-call"}
					}
				}
			]
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))
	resourceManager := resource.NewManager()

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, mockTm, nil, resourceManager, false)
	assert.NoError(t, err)

	// Resource should not be registered
	_, ok := resourceManager.GetResource("http://res1")
	assert.False(t, ok)
}
