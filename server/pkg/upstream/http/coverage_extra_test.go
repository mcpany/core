// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"errors"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"net/http"
	"net/http/httptest"

	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestRegister_Errors(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	t.Run("nil_http_service", func(t *testing.T) {
		configJSON := `{"name": "test-service"}`
		serviceConfig := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.ErrorContains(t, err, "http service config is nil")
	})

	t.Run("empty_address", func(t *testing.T) {
		configJSON := `{
			"name": "test-service",
			"http_service": {
				"address": ""
			}
		}`
		serviceConfig := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.ErrorContains(t, err, "http service address is required")
	})

	t.Run("invalid_address_format", func(t *testing.T) {
		configJSON := `{
			"name": "test-service",
			"http_service": {
				"address": ":invalid"
			}
		}`
		serviceConfig := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.ErrorContains(t, err, "invalid http service address")
	})

	t.Run("invalid_scheme", func(t *testing.T) {
		configJSON := `{
			"name": "test-service",
			"http_service": {
				"address": "ftp://example.com"
			}
		}`
		serviceConfig := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.ErrorContains(t, err, "invalid http service address scheme")
	})
}

func TestRegister_NewHTTPPoolError(t *testing.T) {
	// Mock NewHTTPPool
	originalNewHTTPPool := NewHTTPPool
	defer func() { NewHTTPPool = originalNewHTTPPool }()

	NewHTTPPool = func(minSize, maxSize int, idleTimeout time.Duration, config *configv1.UpstreamServiceConfig) (pool.Pool[*client.HTTPClientWrapper], error) {
		return nil, errors.New("mock pool error")
	}

	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "test-service",
		"http_service": {
			"address": "http://example.com"
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	assert.ErrorContains(t, err, "failed to create HTTP pool")
}

func TestHttpMethodToString_Error(t *testing.T) {
	// Passing an invalid enum value (e.g. -1 or a value not defined in switch)
	// Protobuf enums are int32.
	_, err := httpMethodToString(configv1.HttpCallDefinition_HttpMethod(999))
	assert.ErrorContains(t, err, "unsupported HTTP method")
}

func TestCoverage_ToolDefinitionNotFound(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	// Call defined but no tool definition, and auto-discovery disabled.
	configJSON := `{
		"name": "tool-not-found-service",
		"auto_discover_tool": false,
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [],
			"calls": {
				"c1": {
					"id": "c1",
					"method": "HTTP_METHOD_GET"
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)
	assert.Empty(t, discoveredTools)
}

func TestCoverage_CheckHealth_Failed(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	u := NewUpstream(pm)

	// Use a port that is unlikely to be open, and define health check to force use of checker
	configJSON := `{
		"name": "health-fail",
		"http_service": {
			"address": "http://127.0.0.1:54322",
			"health_check": {
				"expected_code": 200
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := u.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	// Check health. It uses mcphealth.Checker which checks connection because health_check is defined.
	// Since port is closed, it should return error wrapped in "health check failed".
	checker, ok := u.(interface{ CheckHealth(context.Context) error })
	require.True(t, ok)
	err = checker.CheckHealth(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "health check failed")
}

func TestCoverage_CheckHealth_Success(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	u := NewUpstream(pm)

	// We need a real server for health check to pass
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	configJSON := `{
		"name": "health-success",
		"http_service": {
			"address": "` + server.URL + `",
			"health_check": {
				"expected_code": 200,
				"url": "` + server.URL + `"
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := u.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	checker, ok := u.(interface{ CheckHealth(context.Context) error })
	require.True(t, ok)
	err = checker.CheckHealth(context.Background())
	assert.NoError(t, err)
}

func TestCoverage_URLConstruction_ComplexDoubleSlash(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	// //user@host/path -> should be parsed as path //user@host/path
	// This hits the endpointURL.User != nil block

	configJSON := `{
		"name": "double-slash-complex",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [
				{"name": "op1", "call_id": "c1"}
			],
			"calls": {
				"c1": {
					"id": "c1",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "//user@host/path"
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)
	require.Len(t, discoveredTools, 1)

	// Expected: http://127.0.0.1//user@host/path
	fqn := tm.ListTools()[0].Tool().GetUnderlyingMethodFqn()
	assert.Contains(t, fqn, "//user@host/path")
}

func TestCoverage_URLConstruction_ComplexDoubleSlash_Encoded(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	// //user@host/path%2Ffoo
	configJSON := `{
		"name": "double-slash-encoded",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [
				{"name": "op1", "call_id": "c1"}
			],
			"calls": {
				"c1": {
					"id": "c1",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "//user@host/path%2Ffoo"
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)
	require.Len(t, discoveredTools, 1)

	fqn := tm.ListTools()[0].Tool().GetUnderlyingMethodFqn()
	// It should preserve encoding in raw path
	assert.Contains(t, fqn, "//user@host/path%2Ffoo")
}

func TestRegister_PoolConfig_IdleTimeout(t *testing.T) {
	// Mock NewHTTPPool
	originalNewHTTPPool := NewHTTPPool
	defer func() { NewHTTPPool = originalNewHTTPPool }()

	var capturedIdleTimeout time.Duration
	NewHTTPPool = func(minSize, maxSize int, idleTimeout time.Duration, config *configv1.UpstreamServiceConfig) (pool.Pool[*client.HTTPClientWrapper], error) {
		capturedIdleTimeout = idleTimeout
		// Return error to stop execution but we captured what we needed
		return nil, errors.New("stop here")
	}

	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "pool-timeout",
		"http_service": {
			"address": "http://127.0.0.1"
		},
		"connection_pool": {
			"idle_timeout": "60s"
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)

	assert.Equal(t, 60*time.Second, capturedIdleTimeout)
}

func TestCoverage_InputSchema_RequiredNonString(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	// input_schema.required contains non-string.
	configJSON := `{
		"name": "required-non-string",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [
				{"name": "op1", "call_id": "c1"}
			],
			"calls": {
				"c1": {
					"id": "c1",
					"method": "HTTP_METHOD_GET",
					"input_schema": {
						"type": "object",
						"required": ["valid", 123]
					},
					"parameters": [
						{
							"schema": {
								"name": "added",
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

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	tools := tm.ListTools()
	require.Len(t, tools, 1)

	schema := tools[0].Tool().GetAnnotations().GetInputSchema()
	reqVal := schema.Fields["required"].GetListValue()

	// Should contain "valid" and "added". 123 should be filtered out.
	var reqs []string
	for _, v := range reqVal.Values {
		reqs = append(reqs, v.GetStringValue())
	}
	assert.Contains(t, reqs, "valid")
	assert.Contains(t, reqs, "added")
	assert.NotContains(t, reqs, "123") // As string representation
	assert.Len(t, reqs, 2)
}

func TestNewHTTPPool_Coverage(t *testing.T) {
	// We want to test NewHTTPPool logic directly or via Register without mocking.

	t.Run("Resilience Timeout and Private Network", func(t *testing.T) {
		t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", "true")

		pm := pool.NewManager()
		tm := tool.NewManager(nil)
		upstream := NewUpstream(pm)

		configJSON := `{
			"name": "resilience-timeout",
			"http_service": {
				"address": "http://127.0.0.1"
			},
			"resilience": {
				"timeout": "10s"
			}
		}`
		serviceConfig := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		require.NoError(t, err)
	})

	t.Run("Loopback Resources", func(t *testing.T) {
		t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")

		pm := pool.NewManager()
		tm := tool.NewManager(nil)
		upstream := NewUpstream(pm)

		configJSON := `{
			"name": "loopback-resources",
			"http_service": {
				"address": "http://127.0.0.1"
			}
		}`
		serviceConfig := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		require.NoError(t, err)
	})
}

func TestCoverage_MTLS_MissingCA(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	tmpDir := t.TempDir()
	// Use helper from coverage_enhancement_test.go
	_, certPath, keyPath := generateCertFiles(t, tmpDir)

	configJSON := `{
		"name": "mtls-missing-ca",
		"http_service": {
			"address": "https://127.0.0.1"
		},
		"upstream_auth": {
			"mtls": {
				"client_cert_path": "` + certPath + `",
				"client_key_path": "` + keyPath + `",
				"ca_cert_path": "/non/existent/ca.pem"
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create HTTP pool")
	// Verify it failed at CA reading? os.ReadFile usually returns PathError.
	// But we mock NewHTTPPool?
	// NO! TestCoverage_MTLS_MissingCA uses the REAL NewHTTPPool.
	// But wait, TestRegister_NewHTTPPoolError replaced it?
	// And TestRegister_PoolConfig_IdleTimeout replaced it?

	// They deferred restore: defer func() { NewHTTPPool = originalNewHTTPPool }()
	// So it should be restored.
}
