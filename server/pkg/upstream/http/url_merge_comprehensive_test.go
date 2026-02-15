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

func TestHTTPUpstream_URLConstruction_Comprehensive(t *testing.T) {
	testCases := []struct {
		name          string
		address       string
		endpointPath  string
		expectedFqn   string
	}{
		// Query Parameter Merging
		{
			name:         "Merge: Base and Endpoint have different params",
			address:      "http://example.com/api?base=1",
			endpointPath: "/v1/test?endpoint=2",
			expectedFqn:  "GET http://example.com/api/v1/test?base=1&endpoint=2",
		},
		{
			name:         "Merge: Endpoint overrides Base param",
			address:      "http://example.com/api?common=base",
			endpointPath: "/v1/test?common=endpoint",
			expectedFqn:  "GET http://example.com/api/v1/test?common=endpoint",
		},
		{
			name:         "Merge: Endpoint adds value to Base param (same key)",
			// Current implementation seems to override if key matches.
			// Let's verify if it appends or overrides. Based on code reading:
			// if parts, ok := endPartsByKey[bp.key]; ok { if !keysOverridden[bp.key] { finalParts = append(finalParts, parts...) ... } }
			// So it overrides (replaces base values with ALL endpoint values for that key).
			address:      "http://example.com/api?list=1",
			endpointPath: "/v1/test?list=2&list=3",
			expectedFqn:  "GET http://example.com/api/v1/test?list=2&list=3",
		},
		{
			name:         "Merge: Base has multiple values, Endpoint overrides",
			address:      "http://example.com/api?list=1&list=2",
			endpointPath: "/v1/test?list=3",
			expectedFqn:  "GET http://example.com/api/v1/test?list=3",
		},
		{
			name:         "Merge: Base has query, Endpoint has none",
			address:      "http://example.com/api?base=1",
			endpointPath: "/v1/test",
			expectedFqn:  "GET http://example.com/api/v1/test?base=1",
		},
		{
			name:         "Merge: Base has no query, Endpoint has query",
			address:      "http://example.com/api",
			endpointPath: "/v1/test?endpoint=1",
			expectedFqn:  "GET http://example.com/api/v1/test?endpoint=1",
		},
		{
			name:         "Merge: Both have no query",
			address:      "http://example.com/api",
			endpointPath: "/v1/test",
			expectedFqn:  "GET http://example.com/api/v1/test",
		},

		// Special Characters & Encoding
		{
			name:         "Encoding: Space in Base preserved",
			address:      "http://example.com/api?q=hello%20world",
			endpointPath: "/v1/test",
			expectedFqn:  "GET http://example.com/api/v1/test?q=hello%20world",
		},
		{
			name:         "Encoding: Space in Endpoint preserved",
			address:      "http://example.com/api",
			endpointPath: "/v1/test?q=hello%20world",
			expectedFqn:  "GET http://example.com/api/v1/test?q=hello%20world",
		},
		{
			name:         "Encoding: Invalid % in Base preserved",
			address:      "http://example.com/api?q=100%",
			endpointPath: "/v1/test",
			expectedFqn:  "GET http://example.com/api/v1/test?q=100%",
		},
		{
			name:         "Encoding: Invalid % in Endpoint preserved",
			address:      "http://example.com/api",
			endpointPath: "/v1/test?q=100%",
			expectedFqn:  "GET http://example.com/api/v1/test?q=100%",
		},
		{
			name:         "Encoding: Plus sign handling",
			address:      "http://example.com/api?q=a+b",
			endpointPath: "/v1/test",
			expectedFqn:  "GET http://example.com/api/v1/test?q=a+b",
		},

		// Path Resolution
		{
			name:         "Path: Relative path with ..",
			address:      "http://example.com/api/v1/",
			endpointPath: "../v2/test",
			expectedFqn:  "GET http://example.com/api/v2/test",
		},
		{
			name:         "Path: Relative path with .",
			address:      "http://example.com/api/v1/",
			endpointPath: "./test",
			expectedFqn:  "GET http://example.com/api/v1/test",
		},
		{
			name:         "Path: Scheme-relative path (//) treated as path",
			address:      "http://example.com/api",
			endpointPath: "//double/slash",
			expectedFqn:  "GET http://example.com/api//double/slash",
		},
		{
			name:         "Path: Empty endpoint path",
			address:      "http://example.com/api/v1",
			endpointPath: "",
			expectedFqn:  "GET http://example.com/api/v1",
		},
		{
			name:         "Path: Root endpoint path",
			address:      "http://example.com/api",
			endpointPath: "/",
			expectedFqn:  "GET http://example.com/api/",
		},
		{
			name:         "Path: Base has file, Endpoint appends to it (code forces directory behavior)",
			address:      "http://example.com/api/index.html",
			endpointPath: "other.html",
			expectedFqn:  "GET http://example.com/api/index.html/other.html",
		},
		{
			name:         "Path: Base has file, Endpoint relative to it (code forces directory behavior)",
			address:      "http://example.com/api/index.html",
			endpointPath: "./other.html",
			expectedFqn:  "GET http://example.com/api/index.html/other.html",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pm := pool.NewManager()
			tm := tool.NewManager(nil)
			upstream := NewUpstream(pm)

			configJSON := `{
				"name": "url-comprehensive-service",
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
