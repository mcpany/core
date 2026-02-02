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
)

func strPtr(s string) *string {
	return &s
}

func TestHTTPUpstream_URLEdgeCases(t *testing.T) {
	testCases := []struct {
		name         string
		address      string
		endpointPath string
		expectedFqn  string
	}{
		{
			name:         "Invalid key in endpoint query (should be appended)",
			address:      "http://example.com/api?base=1",
			endpointPath: "/test?%ZZ=val", // %ZZ is invalid encoding
			// Expectation: Base preserved, invalid key appended (no override logic for invalid keys)
			expectedFqn: "GET http://example.com/api/test?base=1&%ZZ=val",
		},
		{
			name:         "Double slash at start of endpoint path (should be treated as path)",
			address:      "http://example.com/base",
			endpointPath: "//foo/bar",
			// Expectation: treated as relative path segment "foo/bar" (or similar), NOT scheme-relative
			expectedFqn: "GET http://example.com/base//foo/bar",
		},
		{
			name:         "Empty endpoint path with base trailing slash",
			address:      "http://example.com/api/",
			endpointPath: "",
			expectedFqn:  "GET http://example.com/api/",
		},
		{
			name:         "Empty endpoint path without base trailing slash",
			address:      "http://example.com/api",
			endpointPath: "",
			expectedFqn:  "GET http://example.com/api",
		},
		{
			name:         "Slash endpoint path without base trailing slash",
			address:      "http://example.com/api",
			endpointPath: "/",
			expectedFqn:  "GET http://example.com/api/",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pm := pool.NewManager()
			tm := tool.NewManager(nil)
			upstream := NewUpstream(pm)

			method := configv1.HttpCallDefinition_HTTP_METHOD_GET
			serviceConfig := configv1.UpstreamServiceConfig_builder{
				Name: strPtr("url-edge-cases-service"),
				HttpService: configv1.HttpUpstreamService_builder{
					Address: strPtr(tc.address),
					Tools: []*configv1.ToolDefinition{
						configv1.ToolDefinition_builder{
							Name:   strPtr("test-op"),
							CallId: strPtr("test-op-call"),
						}.Build(),
					},
					Calls: map[string]*configv1.HttpCallDefinition{
						"test-op-call": configv1.HttpCallDefinition_builder{
							Id:           strPtr("test-op-call"),
							Method:       &method,
							EndpointPath: strPtr(tc.endpointPath),
						}.Build(),
					},
				}.Build(),
			}.Build()

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
