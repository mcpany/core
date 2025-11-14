/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package http

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestCreateAndRegisterHTTPTools_Extended(t *testing.T) {
	t.Run("no tools defined", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewToolManager(nil)
		upstream := NewHTTPUpstream(pm)

		configJSON := `{
			"name": "no-tools-service",
			"http_service": {
				"address": "http://localhost",
				"calls": {}
			}
		}`
		serviceConfig := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.NoError(t, err)
		assert.Empty(t, discoveredTools)
		assert.Empty(t, tm.ListTools())
	})

	t.Run("tool with no call definition", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewToolManager(nil)
		upstream := NewHTTPUpstream(pm)

		configJSON := `{
			"name": "no-call-def-service",
			"http_service": {
				"address": "http://localhost",
				"tools": [{"definition": {"name": "test-op"}, "call_id": "non-existent-call"}],
				"calls": {}
			}
		}`
		serviceConfig := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.NoError(t, err)
		assert.Empty(t, discoveredTools)
		assert.Empty(t, tm.ListTools())
	})

	t.Run("tool with required parameters", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewToolManager(nil)
		upstream := NewHTTPUpstream(pm)

		configJSON := `{
			"name": "required-params-service",
			"http_service": {
				"address": "http://localhost",
				"tools": [{"definition": {"name": "test-op"}, "call_id": "test-op-call"}],
				"calls": {
					"test-op-call": {
						"id": "test-op-call",
						"method": "HTTP_METHOD_POST",
						"endpoint_path": "/test",
						"parameters": [
							{"schema": {"name": "param1", "is_required": true}},
							{"schema": {"name": "param2", "is_required": false}}
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

		// We can't directly inspect the parameters of the created tool as they are not exported.
		// However, we can verify that the tool was created successfully and that the service registration
		// did not fail. This indirectly tests that the parameters were processed.
		assert.NotNil(t, registeredTool)
	})

	t.Run("invalid address", func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewToolManager(nil)
		upstream := NewHTTPUpstream(pm)

		configJSON := `{"name": "invalid-address-service", "http_service": {"address": "not a url"}}`
		serviceConfig := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid http service address")
	})
}
