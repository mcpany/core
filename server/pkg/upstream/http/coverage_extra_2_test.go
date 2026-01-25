// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestCoverage_UnsupportedHTTPMethod(t *testing.T) {
	pm := pool.NewManager()
	mockTm := newMockToolManager()
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "unsupported-method-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [
				{"name": "op1", "call_id": "c1"}
			],
			"calls": {
				"c1": {
                    "id": "c1",
                    "method": "HTTP_METHOD_UNSPECIFIED"
                }
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, mockTm, nil, nil, false)
	require.NoError(t, err)

    // Tool should be skipped because method is unsupported
    assert.Empty(t, discoveredTools)
}

func TestCoverage_InvalidEndpointQueryKey(t *testing.T) {
	pm := pool.NewManager()
	mockTm := newMockToolManager()
	upstream := NewUpstream(pm)

    // %GG is invalid key
	configJSON := `{
		"name": "invalid-query-key-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [
				{"name": "op1", "call_id": "c1"}
			],
			"calls": {
				"c1": {
                    "id": "c1",
                    "method": "HTTP_METHOD_GET",
                    "endpoint_path": "/test?%GG=val"
                }
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, mockTm, nil, nil, false)
	require.NoError(t, err)

    require.NotEmpty(t, discoveredTools)
    tools := mockTm.ListTools()
    require.Len(t, tools, 1)

    fqn := tools[0].Tool().GetUnderlyingMethodFqn()
    // It should include the invalid key as is
    assert.Contains(t, fqn, "%GG=val")
}
