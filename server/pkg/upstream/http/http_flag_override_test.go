package http

import (
	"context"
	"net/url"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestHTTPUpstream_URLConstruction_FlagOverride(t *testing.T) {
	// This test verifies that query parameter flag style (presence vs absence of equals sign)
	// is correctly inherited or overridden when merging base URL and endpoint URL.
	// Specifically, it ensures that an endpoint can override a flag parameter from the base URL
	// with an empty-value parameter (e.g., ?flag overrides ?flag=).

	testCases := []struct {
		name          string
		address       string
		endpointPath  string
		expectedQuery string
	}{
		{
			name:          "endpoint overrides flag with empty value",
			address:       "http://example.com/api?flag",
			endpointPath:  "/v1/test?flag=",
			expectedQuery: "flag=",
		},
		{
			name:          "endpoint overrides empty value with flag",
			address:       "http://example.com/api?flag=",
			endpointPath:  "/v1/test?flag",
			expectedQuery: "flag",
		},
		{
			name:          "base has flag, endpoint absent",
			address:       "http://example.com/api?flag",
			endpointPath:  "/v1/test",
			expectedQuery: "flag",
		},
		{
			name:          "base has empty value, endpoint absent",
			address:       "http://example.com/api?flag=",
			endpointPath:  "/v1/test",
			expectedQuery: "flag=",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pm := pool.NewManager()
			tm := tool.NewManager(nil)
			upstream := NewUpstream(pm)

			configJSON := `{
				"name": "flag-bug-service",
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

			toolName := "test-op"
			sanitizedToolName, _ := util.SanitizeToolName(toolName)
			toolID := serviceID + "." + sanitizedToolName
			registeredTool, ok := tm.GetTool(toolID)
			require.True(t, ok, "Tool should be registered")

			httpTool := registeredTool.(*tool.HTTPTool)
			// We can inspect the UnderlyingMethodFqn to see the constructed URL
			def := httpTool.Tool()
			methodFqn := def.GetUnderlyingMethodFqn()

			// format is "GET http://example.com/api/v1/test?..."
			u, err := url.Parse(methodFqn[4:]) // Skip "GET "
			require.NoError(t, err)

			// We check the RawQuery directly
			// The order of params might vary if there are multiple, but here we only have one 'flag'.
			// However, ResolveReference merges queries.
			assert.Equal(t, tc.expectedQuery, u.RawQuery)
		})
	}
}
