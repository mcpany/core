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

func TestHTTPUpstream_ReloadWithRename(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	// Initial registration: Service Name "service-A"
	configJSON1 := `{
		"name": "service-A",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "opA", "call_id": "callA"}],
			"calls": {
				"callA": {
					"id": "callA",
					"method": "HTTP_METHOD_GET"
				}
			}
		}
	}`
	serviceConfig1 := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON1), serviceConfig1))

	idA, _, _, err := upstream.Register(context.Background(), serviceConfig1, tm, nil, nil, false)
	require.NoError(t, err)
	assert.Equal(t, "service-A", idA)

	// Verify tools for service A exist
	toolsA := tm.ListTools()
	assert.Len(t, toolsA, 1)
	assert.Equal(t, "service-A", toolsA[0].Tool().GetServiceId())

	// Reload with NEW Service Name "service-B"
	// This simulates the user changing the "name" field in the config file.
	configJSON2 := `{
		"name": "service-B",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "opB", "call_id": "callB"}],
			"calls": {
				"callB": {
					"id": "callB",
					"method": "HTTP_METHOD_GET"
				}
			}
		}
	}`
	serviceConfig2 := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON2), serviceConfig2))

	// Pass isReload=true
	idB, _, _, err := upstream.Register(context.Background(), serviceConfig2, tm, nil, nil, true)
	require.NoError(t, err)
	assert.Equal(t, "service-B", idB)

	// Verify tools.
	// Expected behavior: Service A tools are gone. Service B tools are present.
	// Actual behavior (suspected): Service A tools remain. Service B tools are added.

	tools := tm.ListTools()

	// Check if service-A tools are still there
	var foundA, foundB bool
	for _, t := range tools {
		if t.Tool().GetServiceId() == "service-A" {
			foundA = true
		}
		if t.Tool().GetServiceId() == "service-B" {
			foundB = true
		}
	}

	assert.True(t, foundB, "Service B tools should be present")
	assert.False(t, foundA, "Service A tools should be removed after rename/reload")
}
