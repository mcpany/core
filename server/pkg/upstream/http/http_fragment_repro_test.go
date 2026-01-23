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

func TestHTTPUpstream_URLConstruction_FragmentRemovalBug(t *testing.T) {
	testCases := []struct {
		name          string
		address       string
		endpointPath  string
		expectedFqn   string
	}{
		{
			name:         "endpoint with new fragment overrides base fragment",
			address:      "http://example.com/api#base",
			endpointPath: "/test#new",
			expectedFqn:  "GET http://example.com/api/test#new",
		},
		{
			name:         "endpoint with no fragment preserves base fragment",
			address:      "http://example.com/api#base",
			endpointPath: "/test",
			expectedFqn:  "GET http://example.com/api/test#base",
		},
		{
			name:         "endpoint with empty fragment should remove base fragment",
			address:      "http://example.com/api#base",
			endpointPath: "/test#",
			// Currently, the code treats empty fragment as "not specified" and restores the base fragment.
			// But if the user explicitly provided #, it implies they want an empty fragment (or no fragment).
			// Ideally, this should be "http://example.com/api/test" or "http://example.com/api/test#"
			// But likely we want to strip it if it's empty? Or just treat empty fragment as valid empty string.
			// Let's assume the user wants to remove the fragment.
			expectedFqn:  "GET http://example.com/api/test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pm := pool.NewManager()
			tm := tool.NewManager(nil)
			upstream := NewUpstream(pm)

			configJSON := `{
				"name": "fragment-test-service",
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
