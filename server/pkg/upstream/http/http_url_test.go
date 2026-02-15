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

func TestHTTPUpstream_URLConstruction_BugRepro(t *testing.T) {
	testCases := []struct {
		name          string
		address       string
		endpointPath  string
		expectedFqn   string
	}{
		{
			name:         "base url with query params should be preserved",
			address:      "http://127.0.0.1:8080?apikey=123",
			endpointPath: "/api/v1/test",
			expectedFqn:  "GET http://127.0.0.1:8080/api/v1/test?apikey=123",
		},
		{
			name:         "base url query and endpoint query should merge",
			address:      "http://127.0.0.1:8080?apikey=123",
			endpointPath: "/api/v1/test?limit=10",
			// Note: order of query params is implementation specific, but typically sorted or preserved.
			// However, in our fix we should ensure both are present.
			// Let's assert that it contains both.
			expectedFqn:  "GET http://127.0.0.1:8080/api/v1/test?apikey=123&limit=10",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pm := pool.NewManager()
			tm := tool.NewManager(nil)
			upstream := NewUpstream(pm)

			configJSON := `{
				"name": "url-test-service",
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

			// For the merge case, we might need flexible matching because map iteration order is random,
			// although url.Values.Encode() sorts by key.
			// "apikey=123&limit=10" is sorted alphabetically.

			assert.Equal(t, tc.expectedFqn, actualFqn)
		})
	}
}
