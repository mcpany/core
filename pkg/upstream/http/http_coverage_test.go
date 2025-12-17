// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestHTTPUpstream_Register_CallPolicy_Blocked(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "blocked-service",
		"http_service": {
			"address": "http://localhost",
			"tools": [
				{"name": "allowed", "call_id": "call-allowed"},
				{"name": "blocked", "call_id": "call-blocked"},
				{"name": "allowed-arg", "call_id": "call-allowed-arg"}
			],
			"calls": {
				"call-allowed": {"id": "call-allowed", "method": "HTTP_METHOD_GET"},
				"call-blocked": {"id": "call-blocked", "method": "HTTP_METHOD_GET"},
				"call-allowed-arg": {"id": "call-allowed-arg", "method": "HTTP_METHOD_GET"}
			}
		},
		"call_policies": [
			{
				"rules": [
					{"name_regex": "blocked", "action": "DENY"},
					{"name_regex": "allowed-arg", "argument_regex": ".*", "action": "DENY"}
				],
				"default_action": "ALLOW"
			}
		]
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	assert.Len(t, discoveredTools, 2)
	assert.Equal(t, "allowed", discoveredTools[0].GetName())
	// allowed-arg should be ALLOWED because argument_regex prevents match at registration time
	assert.Equal(t, "allowed-arg", discoveredTools[1].GetName())
}

func TestHTTPUpstream_Register_MalformedURL(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "bad-url-service",
		"http_service": {
			"address": "http://localhost",
			"tools": [{"name": "bad-op", "call_id": "bad-op-call"}],
			"calls": {
				"bad-op-call": {
					"id": "bad-op-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": ":/bad/path"
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)
	assert.Len(t, discoveredTools, 0, "Tool with invalid URL should be skipped")
}

func TestHTTPUpstream_Register_ExportPolicy(t *testing.T) {
	// Test coverage for "Export Policy"
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "export-policy-test",
		"http_service": {
			"address": "http://localhost",
			"tools": [
				{"name": "private-tool", "call_id": "call1"},
				{"name": "public-tool", "call_id": "call2"}
			],
			"calls": {
				"call1": {"id": "call1", "method": "HTTP_METHOD_GET"},
				"call2": {"id": "call2", "method": "HTTP_METHOD_GET"}
			}
		},
		"tool_export_policy": {
			"rules": [
				{"name_regex": "public.*", "action": "EXPORT"}
			],
			"default_action": "UNEXPORT"
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)
	assert.Len(t, discoveredTools, 1)
	assert.Equal(t, "public-tool", discoveredTools[0].GetName())
}

func TestHTTPUpstream_Register_AutoDiscover(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "autodiscover-test",
		"http_service": {
			"address": "http://localhost",
			"calls": {
				"call1": {"id": "call1", "method": "HTTP_METHOD_GET"},
				"call2": {"id": "call2", "method": "HTTP_METHOD_GET"}
			}
		},
		"auto_discover_tool": true
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)
	assert.Len(t, discoveredTools, 2)
	names := []string{discoveredTools[0].GetName(), discoveredTools[1].GetName()}
	assert.Contains(t, names, "call1")
	assert.Contains(t, names, "call2")
}

func TestHTTPUpstream_Register_Comprehensive(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "comprehensive-test",
		"http_service": {
			"address": "http://localhost",
			"tools": [
				{"name": "allowed-tool", "call_id": "call-allowed"},
				{"name": "blocked-tool", "call_id": "call-blocked"}
			],
			"calls": {
				"call-allowed": {"id": "call-allowed", "method": "HTTP_METHOD_GET"},
				"call-blocked": {"id": "call-blocked", "method": "HTTP_METHOD_GET"},
				"call-dynamic": {"id": "call-dynamic", "method": "HTTP_METHOD_GET"}
			},
			"resources": [
				{"name": "res-private", "uri": "http://private", "disable": false},
				{"name": "res-public", "uri": "http://public", "disable": false},
				{"name": "res-dynamic-orphan", "uri": "http://dynamic", "dynamic": {"http_call": {"id": "call-missing"}}},
				{"name": "res-dynamic-no-tool", "uri": "http://dynamic2", "dynamic": {"http_call": {"id": "call-dynamic"}}}
			],
			"prompts": [
				{"name": "prompt-private", "disable": false},
				{"name": "prompt-public", "disable": false}
			]
		},
		"call_policies": [
			{
				"rules": [
					{"name_regex": "allowed.*", "action": "ALLOW"}
				],
				"default_action": "DENY"
			}
		],
		"resource_export_policy": {
			"rules": [
				{"name_regex": "res-public", "action": "EXPORT"}
			],
			"default_action": "UNEXPORT"
		},
		"prompt_export_policy": {
			"rules": [
				{"name_regex": "prompt-public", "action": "EXPORT"}
			],
			"default_action": "UNEXPORT"
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	// We need to register tool for res-dynamic-no-tool to verify "Tool not found for dynamic resource" check?
	// Actually "call-dynamic" exists but no tool uses it in "tools".
	// So createAndRegisterHTTPTools will NOT create a tool for "call-dynamic" (unless autodiscover).
	// So "tool, ok := callIDToName[call.GetId()]" might find "call-dynamic" -> logic?
	// Wait, callIDToName maps call_id -> tool_name from definitions.
	// "call-dynamic" is NOT in "tools". So it won't be in callIDToName.
	// So "tool not found for dynamic resource" path (line 364) should be hit.

	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
	require.NoError(t, err)

	// Check Tools (Call Policy Default Deny)
	assert.Len(t, discoveredTools, 1)
	assert.Equal(t, "allowed-tool", discoveredTools[0].GetName())

	// Check Resources (Export Policy)
	_, ok := resourceManager.GetResource("http://private")
	assert.False(t, ok, "Private resource should be unexported")
	_, ok = resourceManager.GetResource("http://public")
	assert.True(t, ok, "Public resource should be exported")

	// Check Prompts (Export Policy)
	// Prompt IDs are serviceID.promptName
	// sanitize("comprehensive-test") -> "comprehensive-test"
	serviceID := "comprehensive-test" // sanitize("comprehensive-test")

	// Wait, Register sets u.serviceID = sanitizedName.
	// Let's rely on sanitized logic.
	_, ok = promptManager.GetPrompt(serviceID + ".prompt-private")
	assert.False(t, ok, "Private prompt should be unexported")

	_, ok = promptManager.GetPrompt(serviceID + ".prompt-public")
	assert.True(t, ok, "Public prompt should be exported")
}

func TestHTTPUpstream_Register_InputSchemaGeneration(t *testing.T) {
	// Test automatic input schema generation when InputSchema is missing but Parameters exist
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	// Create parameters manually to ensure they are valid for schemaconv
	// We need configv1.Parameter which maps to jsonschema
	// Actually schemaconv takes []*configv1.Parameter

	configJSON := `{
		"name": "schema-gen-test",
		"http_service": {
			"address": "http://localhost",
			"tools": [
				{"name": "schema-tool", "call_id": "call1"}
			],
			"calls": {
				"call1": {
					"id": "call1",
					"method": "HTTP_METHOD_GET",
					"parameters": [
						{
							"schema": {
								"name": "param1",
								"type": "STRING",
								"is_required": true
							}
						}
					]
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	// Verify the registered tool has correct input schema
	// ID is SHA256("schema-gen-test") ... need to be careful with ID.
	// We can list tools.
	tools := tm.ListTools()
	require.Len(t, tools, 1)
	inputSchema := tools[0].Tool().GetAnnotations().GetInputSchema()
	require.NotNil(t, inputSchema)

	fields := inputSchema.GetFields()
	require.Contains(t, fields, "type")
	require.Equal(t, "object", fields["type"].GetStringValue())
	require.Contains(t, fields, "required")
	reqList := fields["required"].GetListValue().GetValues()
	require.Len(t, reqList, 1)
	assert.Equal(t, "param1", reqList[0].GetStringValue())
}

func TestHTTPUpstream_Register_InvalidAddress(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "invalid-address",
		"http_service": {
			"address": ":/invalid"
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid http service address")
}

func TestHTTPUpstream_Register_PoolConfig(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "pool-config-test",
		"http_service": {
			"address": "http://localhost"
		},
		"connection_pool": {
			"max_connections": 5,
			"max_idle_connections": 2,
			"idle_timeout": "10s"
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)
}
