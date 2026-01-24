// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestHTTPUpstream_CheckHealth_WithChecker(t *testing.T) {
	// Start a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	u := NewUpstream(pm)

	// Configure with health check
	configJSON := `{
		"name": "health-check-service",
		"http_service": {
			"address": "` + server.URL + `",
			"health_check": {
				"url": "` + server.URL + `/health",
				"expected_code": 200,
				"expected_response_body_contains": "OK"
			},
			"tools": [{"name": "test-op", "call_id": "test-op-call"}],
			"calls": {
				"test-op-call": {
					"id": "test-op-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/test"
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := u.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	checker, ok := u.(upstream.HealthChecker)
	require.True(t, ok)

	// CheckHealth calls u.checker.Check(ctx).
	err = checker.CheckHealth(context.Background())
	assert.NoError(t, err)

	// Test failing health check
	server.Close() // Make it fail

	// We might need to wait for cache to expire?
	// The checker is configured with WithCacheDuration(1 * time.Second) in health.go.
	time.Sleep(1200 * time.Millisecond)

	err = checker.CheckHealth(context.Background())
	assert.Error(t, err)
}

func TestHTTPUpstream_Register_CoverageEdges(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	u := NewUpstream(pm)

	t.Run("Unsupported Method", func(t *testing.T) {
		configJSON := `{
			"name": "unsupported-method-service",
			"http_service": {
				"address": "http://example.com",
				"tools": [{"name": "test-op", "call_id": "test-op-call"}],
				"calls": {
					"test-op-call": {
						"id": "test-op-call",
						"method": "HTTP_METHOD_UNSPECIFIED",
						"endpoint_path": "/test"
					}
				}
			}
		}`
		serviceConfig := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		_, tools, _, err := u.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.NoError(t, err)
		assert.Empty(t, tools, "Should have 0 tools registered because method was unsupported")
	})

	t.Run("Invalid Endpoint Path Parsing - Manual Struct", func(t *testing.T) {
		badPath := "/test\x7f"
		method := configv1.HttpCallDefinition_HTTP_METHOD_GET

		svcConfig := &configv1.UpstreamServiceConfig{
			Name: proto.String("bad-path-service"),
			ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
				HttpService: &configv1.HttpUpstreamService{
					Address: proto.String("http://example.com"),
					Tools: []*configv1.ToolDefinition{
						{Name: proto.String("test-op"), CallId: proto.String("test-op-call")},
					},
					Calls: map[string]*configv1.HttpCallDefinition{
						"test-op-call": {
							Id: proto.String("test-op-call"),
							Method: &method,
							EndpointPath: proto.String(badPath),
						},
					},
				},
			},
		}

		_, tools, _, err := u.Register(context.Background(), svcConfig, tm, nil, nil, false)
		assert.NoError(t, err)
		assert.Empty(t, tools)
	})

	t.Run("Pool Config Coverage", func(t *testing.T) {
		// Cover the branch where poolConfig is present
		svcConfig := &configv1.UpstreamServiceConfig{
			Name: proto.String("pool-config-service"),
			ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
				HttpService: &configv1.HttpUpstreamService{
					Address: proto.String("http://example.com"),
				},
			},
			ConnectionPool: &configv1.ConnectionPoolConfig{
				MaxConnections: protoInt32(50),
				MaxIdleConnections: protoInt32(5),
				IdleTimeout: durationpb.New(30 * time.Second),
			},
		}

		_, _, _, err := u.Register(context.Background(), svcConfig, tm, nil, nil, false)
		assert.NoError(t, err)
	})

	t.Run("Disabled Tool", func(t *testing.T) {
		configJSON := `{
			"name": "disabled-tool-service",
			"http_service": {
				"address": "http://example.com",
				"tools": [
                    {"name": "enabled-tool", "call_id": "enabled-call"},
                    {"name": "disabled-tool", "call_id": "disabled-call", "disable": true}
                ],
				"calls": {
					"enabled-call": {
						"id": "enabled-call",
						"method": "HTTP_METHOD_GET",
						"endpoint_path": "/enabled"
					},
					"disabled-call": {
						"id": "disabled-call",
						"method": "HTTP_METHOD_GET",
						"endpoint_path": "/disabled"
					}
				}
			}
		}`
		serviceConfig := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		_, tools, _, err := u.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.NoError(t, err)
		assert.Len(t, tools, 1)
		assert.Equal(t, "enabled-tool", tools[0].GetName())
	})
}

func protoInt32(i int32) *int32 {
	return &i
}
