package http

import (
	"context"
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

func TestHTTPUpstream_Coverage_EdgeCases(t *testing.T) {
	t.Run("invalid base url", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewManager(nil)
		upstream := NewUpstream(pm)

		configJSON := `{
			"name": "invalid-url",
			"http_service": {
				"address": "::invalid-url::",
				"tools": [{"name": "test", "call_id": "call"}],
				"calls": {"call": {"id": "call", "method": "HTTP_METHOD_GET"}}
			}
		}`
		serviceConfig := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		// Should error during Register because ValidateHTTPServiceDefinition (if used) or url.Parse would fail.
		// In Register function, it calls url.ParseRequestURI(address).
		_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid http service address")
	})

	t.Run("fail to create pool", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewManager(nil)
		upstream := NewUpstream(pm)

		configJSON := `{
			"name": "fail-pool",
			"http_service": {
				"address": "http://localhost",
				"tools": [{"name": "test", "call_id": "call"}],
				"calls": {"call": {"id": "call", "method": "HTTP_METHOD_GET"}}
			}
		}`
		serviceConfig := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		// Mock NewHTTPPool to fail
		originalNewHTTPPool := NewHTTPPool
		defer func() { NewHTTPPool = originalNewHTTPPool }()
		NewHTTPPool = func(minSize, maxSize int, idleTimeout time.Duration, config *configv1.UpstreamServiceConfig) (pool.Pool[*client.HTTPClientWrapper], error) {
			return nil, assert.AnError
		}

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create HTTP pool")
	})

	t.Run("call policy compile error", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewManager(nil)
		upstream := NewUpstream(pm)

		configJSON := `{
			"name": "policy-compile-error",
			"http_service": {
				"address": "http://localhost",
				"tools": [{"name": "test", "call_id": "call"}],
				"calls": {"call": {"id": "call", "method": "HTTP_METHOD_GET"}}
			},
			"call_policies": [
				{
					"rules": [{"name_regex": "(", "action": "DENY"}]
				}
			]
		}`
		serviceConfig := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		// Register currently returns nil tools if policy compilation fails, but no error (it logs it).
		// We can check that no tools were registered.
		assert.NoError(t, err)
		assert.Empty(t, discoveredTools)
	})

	t.Run("endpoint path parse error", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewManager(nil)
		upstream := NewUpstream(pm)

		// Create a path that fails url.Parse.
		// A control character might work.
		invalidPath := string([]byte{0x7f})

		method := configv1.HttpCallDefinition_HTTP_METHOD_GET
		serviceConfig := &configv1.UpstreamServiceConfig{
			Name: String("endpoint-parse-error"),
			ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
				HttpService: &configv1.HttpUpstreamService{
					Address: String("http://localhost"),
					Tools: []*configv1.ToolDefinition{
						{Name: String("test"), CallId: String("call")},
					},
					Calls: map[string]*configv1.HttpCallDefinition{
						"call": {
							Id:           String("call"),
							Method:       &method,
							EndpointPath: String(invalidPath),
						},
					},
				},
			},
		}

		_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.NoError(t, err)
		assert.Empty(t, discoveredTools)
	})

	t.Run("endpoint path parse error for encoded slash fix", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewManager(nil)
		upstream := NewUpstream(pm)

		// Path starts with %2F but is otherwise invalid for url.Parse("./" + path)
		// This is hard because url.Parse is quite robust.
		// However, if we put a control char after %2F.
		invalidPath := "%2F" + string([]byte{0x7f})

		method := configv1.HttpCallDefinition_HTTP_METHOD_GET
		serviceConfig := &configv1.UpstreamServiceConfig{
			Name: String("encoded-slash-parse-error"),
			ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
				HttpService: &configv1.HttpUpstreamService{
					Address: String("http://localhost"),
					Tools: []*configv1.ToolDefinition{
						{Name: String("test"), CallId: String("call")},
					},
					Calls: map[string]*configv1.HttpCallDefinition{
						"call": {
							Id:           String("call"),
							Method:       &method,
							EndpointPath: String(invalidPath),
						},
					},
				},
			},
		}

		_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.NoError(t, err)
		assert.Empty(t, discoveredTools)
	})

	t.Run("endpoint path parse error for double slash fix", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewManager(nil)
		upstream := NewUpstream(pm)

		// Path starts with // but invalid after trimming
		invalidPath := "//" + string([]byte{0x7f})

		method := configv1.HttpCallDefinition_HTTP_METHOD_GET
		serviceConfig := &configv1.UpstreamServiceConfig{
			Name: String("double-slash-parse-error"),
			ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
				HttpService: &configv1.HttpUpstreamService{
					Address: String("http://localhost"),
					Tools: []*configv1.ToolDefinition{
						{Name: String("test"), CallId: String("call")},
					},
					Calls: map[string]*configv1.HttpCallDefinition{
						"call": {
							Id:           String("call"),
							Method:       &method,
							EndpointPath: String(invalidPath),
						},
					},
				},
			},
		}

		_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.NoError(t, err)
		assert.Empty(t, discoveredTools)
	})
}

func String(s string) *string {
	return &s
}
