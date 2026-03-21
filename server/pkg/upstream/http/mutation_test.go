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
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestInputSchemaMutation(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "mutation-test-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "test-op", "call_id": "test-op-call"}],
			"calls": {
				"test-op-call": {
					"id": "test-op-call",
					"method": "HTTP_METHOD_POST",
					"endpoint_path": "/test",
					"parameters": [
						{
							"schema": {
								"name": "extra_param",
								"type": "STRING",
								"is_required": true
							}
						}
					],
					"input_schema": {
						"type": "object",
						"properties": {
							"original_param": {
								"type": "string"
							}
						}
					}
				}
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	// First registration
	serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	sanitizedToolName, _ := util.SanitizeToolName("test-op")
	toolID := serviceID + "." + sanitizedToolName
	_, ok := tm.GetTool(toolID)
	require.True(t, ok)

	// Check if the input_schema in serviceConfig was mutated
	httpService := serviceConfig.GetHttpService()
	calls := httpService.GetCalls()
	callDef := calls["test-op-call"]
	inputSchema := callDef.GetInputSchema()

	// The "extra_param" should NOT be in the original input_schema in the config object
	// if we are respecting immutability.

	properties := inputSchema.Fields["properties"].GetStructValue().Fields
	if _, ok := properties["extra_param"]; ok {
		t.Fatalf("Bug confirmed: serviceConfig was mutated! 'extra_param' found in input_schema.")
	}
}
