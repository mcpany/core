// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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
			// We expect "flag" to be a flag (no equals).
			// We accept "&" as separator in output because Encode() forces it.
			// But "flag" MUST NOT become "flag=".
			// Note: parseQueryManual reorders/reconstructs.
			// url.Values.Encode() sorts by key. "flag" < "param".
			expectedFqn:  "GET http://example.com/api/test?flag&param=val",
		},
		{
			name:         "endpoint with semicolon separated flags",
			address:      "http://example.com/api",
			endpointPath: "/test?flag1;flag2",
			expectedFqn:  "GET http://example.com/api/test?flag1&flag2",
		},
		{
			name:         "mixed ampersand and semicolon",
			address:      "http://example.com/api",
			endpointPath: "/test?a=1&flag;b=2",
			expectedFqn:  "GET http://example.com/api/test?a=1&b=2&flag",
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
