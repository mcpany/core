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

func TestHTTPUpstream_URLConstruction_EscapedSlashBug(t *testing.T) {
	testCases := []struct {
		name          string
		address       string
		endpointPath  string
		expectedFqn   string
	}{
		{
			name:         "endpoint starting with escaped slash should not have dot prepended",
			address:      "http://example.com/api",
			endpointPath: "%2Ffoo",
			// Decoded path is /foo. Code sees it starts with / and prepends . to RawPath (%2Ffoo).
			// Resulting in .%2Ffoo, which ResolveReference treats literally.
			// Expected: http://example.com/api/%2Ffoo
			expectedFqn:  "GET http://example.com/api/%2Ffoo",
		},
		{
			name:         "endpoint starting with normal slash",
			address:      "http://example.com/api",
			endpointPath: "/foo",
			expectedFqn:  "GET http://example.com/api/foo",
		},
		{
			name:         "endpoint starting with double slash (fix regression check)",
			address:      "http://example.com/api",
			endpointPath: "//foo",
			expectedFqn:  "GET http://example.com/api//foo",
		},
		{
			name:         "endpoint with escaped char but starts with slash",
			address:      "http://example.com/api",
			endpointPath: "/foo%20bar",
			// Path: /foo bar (starts with /). RawPath: /foo%20bar (starts with /).
			// mustPrepend should be true.
			expectedFqn:  "GET http://example.com/api/foo%20bar",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pm := pool.NewManager()
			tm := tool.NewManager(nil)
			upstream := NewUpstream(pm)

			configJSON := `{
				"name": "escaped-slash-service",
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
