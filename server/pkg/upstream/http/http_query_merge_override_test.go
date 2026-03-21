// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestHTTPUpstream_QueryMerge_InvalidBaseParam_Override(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	// Base URL has an invalid query parameter (q=val% which is invalid hex)
	// Endpoint path tries to override 'q' with a valid value.
	configJSON := `{
		"name": "query-merge-test",
		"http_service": {
			"address": "http://example.com?q=val%&other=keep",
			"tools": [{"name": "test-op", "call_id": "test-op-call"}],
			"calls": {
				"test-op-call": {
					"id": "test-op-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/path?q=valid"
				}
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	registeredTool, ok := tm.GetTool(serviceID + ".test-op")
	require.True(t, ok)

	fqn := registeredTool.Tool().GetUnderlyingMethodFqn()
	// We expect q=valid to override q=val%.
	// Current buggy behavior: both are present, or q=val% remains.
	// Expected: GET http://example.com/path?other=keep&q=valid (order might vary)

	assert.Contains(t, fqn, "q=valid")
	assert.NotContains(t, fqn, "q=val%")
	assert.Contains(t, fqn, "other=keep")
}
