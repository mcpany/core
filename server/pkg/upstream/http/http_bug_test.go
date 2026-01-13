package http

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestHTTPUpstream_URLConstruction_EncodedSlashBug(t *testing.T) {
	testCases := []struct {
		name          string
		address       string
		endpointPath  string
		expectedFqn   string
	}{
		{
			name:         "endpoint starting with encoded slash",
			address:      "http://localhost:8080/api",
			endpointPath: "%2Ftest",
			// We expect the encoded slash to be preserved as a path segment
			// resulting in /api//test (where the second slash is the decoded %2F)
			// OR /api/%2Ftest (if we want to preserve encoding).
			// Usually, for an API that expects %2Ftest, it means a segment named "/test".
			// If we append it to /api, it should be /api/%2Ftest.
			expectedFqn:  "GET http://localhost:8080/api/%2Ftest",
		},
		{
			name:         "endpoint starting with encoded slash and trailing slash",
			address:      "http://localhost:8080/api",
			endpointPath: "%2Ftest/",
			expectedFqn:  "GET http://localhost:8080/api/%2Ftest/",
		},
		{
			name:         "endpoint starting with double slash (regression check)",
			address:      "http://localhost:8080/api",
			endpointPath: "//test",
			// Existing logic fixes // to /
			expectedFqn:  "GET http://localhost:8080/api/test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pm := pool.NewManager()
			tm := tool.NewManager(nil)
			upstream := NewUpstream(pm)

			configJSON := `{
				"name": "encoded-slash-service",
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
