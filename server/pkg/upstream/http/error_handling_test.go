package http

import (
	"context"
	"errors"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestHTTPUpstream_ErrorHandling(t *testing.T) {
	t.Run("NewHTTPPool Failure", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewManager(nil)
		upstream := NewUpstream(pm)

		configJSON := `{
			"name": "pool-fail-service",
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
			}
		}`
		serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		// Swizzle NewHTTPPool
		originalNewHTTPPool := NewHTTPPool
		defer func() { NewHTTPPool = originalNewHTTPPool }()

		NewHTTPPool = func(minSize, maxSize int, idleTimeout time.Duration, config *configv1.UpstreamServiceConfig) (pool.Pool[*client.HTTPClientWrapper], error) {
			return nil, errors.New("simulated pool creation failure")
		}

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "simulated pool creation failure")
		assert.Contains(t, err.Error(), "failed to create HTTP pool")
	})

	t.Run("Invalid Service Address", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewManager(nil)
		upstream := NewUpstream(pm)

		// using a control character to force url.ParseRequestURI error
		// We use \u007f in the raw string literal. In JSON, \u007f escapes to the control character.
		configJSON := `{
			"name": "invalid-addr-service",
			"http_service": {
				"address": "http://127.0.0.1/\u007f",
				"tools": [{"name": "test-op", "call_id": "test-op-call"}],
				"calls": {}
			}
		}`
		serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid http service address")
	})

	t.Run("Invalid Endpoint Path", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewManager(nil)
		upstream := NewUpstream(pm)

		configJSON := `{
			"name": "invalid-endpoint-service",
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
		serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.NoError(t, err)
		assert.Len(t, discoveredTools, 0, "Should not register tool with invalid endpoint path")
	})

	t.Run("Invalid HTTP Method", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewManager(nil)
		upstream := NewUpstream(pm)

		configJSON := `{
			"name": "invalid-method-service",
			"http_service": {
				"address": "http://127.0.0.1",
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
		serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.NoError(t, err)
		assert.Len(t, discoveredTools, 0, "Should not register tool with invalid HTTP method")
	})

	t.Run("Call Policy Deny", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewManager(nil)
		upstream := NewUpstream(pm)

		configJSON := `{
			"name": "policy-deny-service",
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
					"default_action": "DENY"
				}
			]
		}`
		serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.NoError(t, err)
		assert.Len(t, discoveredTools, 0, "Should not register tool denied by policy")
	})

	t.Run("Parse Base URL Failure (in loop)", func(t *testing.T) {
		// This is hard to trigger because Register checks the address first.
		// However, createAndRegisterHTTPTools parses it AGAIN.
		// If we can pass Register check but fail loop check?
		// Register uses url.ParseRequestURI. loop uses url.Parse.
		// url.ParseRequestURI is generally stricter (requires scheme).
		// So if Register passes, loop likely passes.
		// But maybe we can modify the config object AFTER Register validation but BEFORE the loop?
		// No, Register calls createAndRegisterHTTPTools immediately.
		//
		// Wait, createAndRegisterHTTPTools parses `address` again.
		// If `Register` does:
		// uURL, err := url.ParseRequestURI(address)
		// And `createAndRegisterHTTPTools` does:
		// baseURL, err := url.Parse(address)
		//
		// Are there strings that pass ParseRequestURI but fail Parse?
		// Probably not.
		// But we can check if `createAndRegisterHTTPTools` handles `nil` return from `url.Parse` if we were to force it (which we can't easily).
		//
		// Actually, we can assume that if `Register` passes, `createAndRegisterHTTPTools`'s parse will likely succeed.
		// So this branch might be unreachable in practice unless there's a race condition or some very weird URL.
	})
}
