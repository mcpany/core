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

func TestHTTPUpstream_InputSchema_BugRepro(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	// Configuration with both InputSchema and Parameters.
	// Parameters define a required path parameter.
	// InputSchema defines a body parameter but omits the path parameter (or doesn't mark it required).
	// We expect the final tool schema to include the required path parameter.
	configJSON := `{
		"name": "schema-bug-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "test-op", "call_id": "test-op-call"}],
			"calls": {
				"test-op-call": {
					"id": "test-op-call",
					"method": "HTTP_METHOD_POST",
					"endpoint_path": "/users/{user_id}",
					"parameters": [
						{
							"schema": {
								"name": "user_id",
								"type": "STRING",
								"is_required": true
							}
						}
					],
					"input_schema": {
						"type": "object",
						"properties": {
							"body_field": {
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

	serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	sanitizedToolName, _ := util.SanitizeToolName("test-op")
	toolID := serviceID + "." + sanitizedToolName
	registeredTool, ok := tm.GetTool(toolID)
	require.True(t, ok)

	inputSchema := registeredTool.Tool().GetAnnotations().GetInputSchema()
	require.NotNil(t, inputSchema)

	fields := inputSchema.GetFields()
	require.Contains(t, fields, "required")

	requiredList := fields["required"].GetListValue().GetValues()
	found := false
	for _, v := range requiredList {
		if v.GetStringValue() == "user_id" {
			found = true
			break
		}
	}
	assert.True(t, found, "user_id should be in required list")

	// Also check if user_id is in properties.
	props := fields["properties"].GetStructValue().GetFields()
	_, propFound := props["user_id"]
	assert.True(t, propFound, "user_id should be in properties")
}
