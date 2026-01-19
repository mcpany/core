// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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

func TestHTTPUpstream_URLConstruction_QueryMergeBug(t *testing.T) {
	testCases := []struct {
		name          string
		address       string
		endpointPath  string
		expectedFqn   string
	}{
		{
			name:         "base url with flag query param should not gain equals sign when merged with endpoint query",
			address:      "http://example.com/api?flag",
			endpointPath: "/v1/test?foo=bar",
			expectedFqn:  "GET http://example.com/api/v1/test?flag&foo=bar",
		},
		{
			name:         "base url with encoded flag (space) should be preserved",
			address:      "http://example.com/api?a%20b",
			endpointPath: "/v1/test?foo=bar",
			// "a%20b" decodes to "a b". Encode() typically outputs "a+b" (or "a%20b").
			// The key is to ensure it doesn't have "=".
			// We accept either "a+b" or "a%20b" as long as it's a valid encoding of "a b" and has no "=".
			// However, since we are reconstructing from encoded parts, we expect it to match the encoder's output for the key,
			// which is usually "a+b" in Go's url package.
			expectedFqn:  "GET http://example.com/api/v1/test?a+b&foo=bar",
		},
		{
			name:         "flag overridden by value should have equals",
			address:      "http://example.com/api?flag",
			endpointPath: "/v1/test?flag=true",
			expectedFqn:  "GET http://example.com/api/v1/test?flag=true",
		},
		{
			name:         "flag overridden by empty value should have equals",
			address:      "http://example.com/api?flag",
			endpointPath: "/v1/test?flag=",
			expectedFqn:  "GET http://example.com/api/v1/test?flag=",
		},
		{
			name:         "non-flag empty value should have equals",
			address:      "http://example.com/api?existing=val",
			endpointPath: "/v1/test?new=",
			expectedFqn:  "GET http://example.com/api/v1/test?existing=val&new=",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pm := pool.NewManager()
			tm := tool.NewManager(nil)
			upstream := NewUpstream(pm)

			configJSON := `{
				"name": "query-merge-service",
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
