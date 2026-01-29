package http

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestHTTPUpstream_URLConstruction_SemicolonBug(t *testing.T) {
	testCases := []struct {
		name          string
		address       string
		endpointPath  string
		expectedFqn   string
	}{
		{
			name:         "endpoint with semicolon separated flag",
			address:      "http://example.com/api",
			endpointPath: "/test?flag;param=val",
			// Semicolon is NO LONGER supported as a separator.
			// "flag;param=val" is treated as a single flag key "flag;param=val".
			// We now preserve the original string if possible, so it stays "flag;param=val".
			expectedFqn:  "GET http://example.com/api/test?flag;param=val",
		},
		{
			name:         "endpoint with semicolon separated flags",
			address:      "http://example.com/api",
			endpointPath: "/test?flag1;flag2",
			// Treated as single flag "flag1;flag2", preserved as is.
			expectedFqn:  "GET http://example.com/api/test?flag1;flag2",
		},
		{
			name:         "mixed ampersand and semicolon",
			address:      "http://example.com/api",
			endpointPath: "/test?a=1&flag;b=2",
			// "a=1" is parsed as key "a" val "1".
			// "flag;b=2" is parsed as flag key "flag;b=2", preserved as is.
			expectedFqn:  "GET http://example.com/api/test?a=1&flag;b=2",
		},
		{
			name:         "semicolon in query value should be preserved (value)",
			address:      "http://example.com/api",
			endpointPath: "/v1/test?q=hello;world",
			// We expect the semicolon to be preserved in the value as is.
			expectedFqn:  "GET http://example.com/api/v1/test?q=hello;world",
		},
		{
			name:         "semicolon in base url query value should be preserved (literal)",
			address:      "http://example.com/api?q=hello;world",
			endpointPath: "/v1/test",
			// Since no merge happens (endpoint has no query), the base query is preserved literally.
			expectedFqn:  "GET http://example.com/api/v1/test?q=hello;world",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pm := pool.NewManager()
			tm := tool.NewManager(nil)
			upstream := NewUpstream(pm)

			configJSON := `{
				"name": "semicolon-service",
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

			actualFqn := registeredTool.Tool().GetUnderlyingMethodFqn()
			assert.Equal(t, tc.expectedFqn, actualFqn)
		})
	}
}
