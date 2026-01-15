// Copyright 2025 Author(s) of MCP Any
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

func TestHTTPUpstream_DefaultValue_E2E(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	// Configuration with a parameter having a default value
	configJSON := `{
		"name": "default-value-service",
		"http_service": {
			"address": "http://localhost",
			"tools": [{"name": "test-op", "call_id": "test-op-call"}],
			"calls": {
				"test-op-call": {
					"id": "test-op-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/test",
					"parameters": [
						{
							"schema": {
								"name": "limit",
								"type": "INTEGER",
								"default_value": 10
							}
						}
					]
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
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
	props := fields["properties"].GetStructValue().GetFields()

	limitProp, ok := props["limit"]
	require.True(t, ok, "limit parameter should be in properties")

	limitFields := limitProp.GetStructValue().GetFields()
	defaultVal, hasDefault := limitFields["default"]

	assert.True(t, hasDefault, "limit parameter should have default value")
	if hasDefault {
		assert.Equal(t, float64(10), defaultVal.GetNumberValue())
	}
}
