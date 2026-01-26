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

func TestHTTPUpstream_DuplicateCallID(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "duplicate-call-id-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [
				{"name": "tool1", "call_id": "call1"},
				{"name": "tool2", "call_id": "call1"}
			],
			"calls": {
				"call1": {
					"id": "call1",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/test"
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	serviceID, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	assert.NoError(t, err)

	t.Logf("Discovered %d tools", len(discoveredTools))
	for _, dt := range discoveredTools {
		t.Logf("Tool: %s", dt.GetName())
	}

	// We verify that both tools are registered, even if they share the same call_id.
	// This is a valid use case: reusing the same underlying API call but exposing it as different tools (maybe with different descriptions or hints, although in this minimal config they are similar).

	sanitizedToolName1, _ := util.SanitizeToolName("tool1")
	toolID1 := serviceID + "." + sanitizedToolName1
	_, ok1 := tm.GetTool(toolID1)

	sanitizedToolName2, _ := util.SanitizeToolName("tool2")
	toolID2 := serviceID + "." + sanitizedToolName2
	_, ok2 := tm.GetTool(toolID2)

	assert.True(t, ok1, "Tool 1 should be registered")
	assert.True(t, ok2, "Tool 2 should be registered")
	assert.Len(t, discoveredTools, 2, "Expected 2 tools to be registered")
}
