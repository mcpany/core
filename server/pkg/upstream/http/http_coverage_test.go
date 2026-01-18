// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/client"

	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
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

func TestHTTPUpstream_Register_PoolCreationFailure(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "pool-fail-service",
		"http_service": {
			"address": "http://localhost"
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	// Mock NewHTTPPool to fail
	originalNewHTTPPool := NewHTTPPool
	defer func() { NewHTTPPool = originalNewHTTPPool }()

	NewHTTPPool = func(_, _ int, _ time.Duration, _ *configv1.UpstreamServiceConfig) (pool.Pool[*client.HTTPClientWrapper], error) {
		return nil, errors.New("mock pool creation failed")
	}

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock pool creation failed")
}

func TestHTTPUpstream_Register_ResourceErrors(t *testing.T) {
	pm := pool.NewManager()
	upstream := NewUpstream(pm)

	// We need a service with tools to test linking
	configJSON := `{
		"name": "resource-error-service",
		"http_service": {
			"address": "http://localhost",
			"tools": [
				{"name": "tool1", "call_id": "call1"}
			],
			"calls": {
				"call1": {"id": "call1", "method": "HTTP_METHOD_GET"}
			},
			"resources": [
				{
					"name": "res-missing-call",
					"uri": "http://res1",
					"dynamic": {}
				},
				{
					"name": "res-unknown-call",
					"uri": "http://res2",
					"dynamic": {"http_call": {"id": "unknown-call"}}
				},
				{
					"name": "res-tool-not-found",
					"uri": "http://res3",
					"dynamic": {"http_call": {"id": "call1"}}
				}
			]
		}
	}`
	// Note: for res-tool-not-found, we need "call1" to be in config, so it maps to "tool1",
	// but we will trick the toolManager to NOT have "tool1" registered, or failing `GetTool`.
	// Actually `Register` calls `createAndRegisterHTTPTools` first, which registers tools.
	// So `tool1` WILL be registered.
	// To trigger "Tool not found for dynamic resource", we might need to simulate `toolManager.GetTool` failure.
	// Or we can use `call-dynamic` which doesn't have a tool definition but is in calls?
	// If it's not in `tools` list, it won't be registered unless auto-discover is on.
	// If it's not in `tools` list, `callIDToName` won't have it. -> "tool not found for dynamic resource" log and continue. (Line 364 in http.go)

	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	// Mock tool manager to control GetTool?
	// The standard `tool.NewManager` works fine.
	// For `res-tool-not-found`:
	// If we provide a call that IS in `tools` (so callIDToName has it),
	// but for some reason `toolManager.GetTool` returns false.
	// `createAndRegisterHTTPTools` adds tools to manager.
	// If we want it to fail look up, maybe we can delete it from manager concurrently? No, standard logic is synchronous.
	// Actually, `callIDToName` is built from `httpService.GetTools()`.
	// `discoveredTools` are added to `toolManager`.
	// If we have a tool in `httpService.GetTools()` but `AddTool` fails?
	// We can use a mock tool manager again if we want precise control.

	// Let's use `NewMockToolManager` from `http_test.go` if we can import/access it?
	// It's in `http_test.go`, same package, but not exported. `http_coverage_test.go` is same package `http`.
	// So we can use `newMockToolManager`.

	mockTm := newMockToolManager()

	// For `res-missing-call`: dynamic with nil http_call. Handled?
	// Proto `GetHttpCall()` returns nil if missing. Code: `if call == nil { continue }`

	// For `res-unknown-call`: `unknown-call` not in `callIDToName`. -> "tool not found for dynamic resource" (id error).

	// For `res-tool-not-found`:
	// We want `call1` -> `tool1` in `callIDToName`.
	// But `toolManager.GetTool(serviceID + ".tool1")` to return false.
	// `createAndRegisterHTTPTools` adds it to `mockTm`.
	// We can make `mockTm` fail `GetTool` for specific name?
	// Or we can make `AddTool` fail so it never gets added?
	// If `AddTool` fails, `discoveredTools` won't have it. `callIDToName` WILL have it (it comes from config).
	// So `callIDToName["call1"]` = "tool1".
	// `toolManager.GetTool` will fail.

	mockTm.addError = errors.New("failed to add tool")
	// This will cause all tools to fail addition.

	rm := resource.NewManager()

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, mockTm, nil, rm, false)
	require.NoError(t, err)

	// Verify resources were skipped
	// "res-missing-call" -> skipped (continue)
	// "res-unknown-call" -> skipped (log error)
	// "res-tool-not-found" -> skipped (log error because tool1 not added)

	assert.Empty(t, rm.ListResources())
}

func TestHTTPUpstream_Register_PromptExportPolicy(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)
	promptManager := prompt.NewManager()

	configJSON := `{
		"name": "prompt-export-test",
		"http_service": {
			"address": "http://localhost",
			"prompts": [
				{"name": "p-default-allow", "disable": false},
				{"name": "p-default-deny", "disable": false},
				{"name": "p-explicit-allow", "disable": false},
				{"name": "p-explicit-deny", "disable": false}
			]
		},
		"prompt_export_policy": {
			"rules": [
				{"name_regex": "p-explicit-allow", "action": "EXPORT"},
				{"name_regex": "p-explicit-deny", "action": "UNEXPORT"}
			],
			"default_action": "UNEXPORT"
		}
	}`
	// Wait, if default is UNEXPORT, p-default-allow (intended) will be unexported.
	// My naming is confusing.

	// Let's test standard case:
	// - p-export: matches EXPORT rule
	// - p-unexport: matches UNEXPORT rule
	// - p-residue: falls invalid default (UNEXPORT)

	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, nil, false)
	require.NoError(t, err)

	// p-explicit-allow -> Should be present
	// p-explicit-deny -> Should be absent
	// p-default-allow/deny -> Should be absent because default is UNEXPORT

	prompts := promptManager.ListPrompts()
	// Depending on earlier tests, promptManager might be empty.
	// Name includes service ID.

	foundAllow := false
	foundDeny := false
	for _, p := range prompts {
		if p.Prompt().Name == "prompt-export-test.p-explicit-allow" {
			foundAllow = true
		}
		if p.Prompt().Name == "prompt-export-test.p-explicit-deny" {
			foundDeny = true
		}
	}
	assert.True(t, foundAllow, "Explicit allow should be exported")
	assert.False(t, foundDeny, "Explicit deny should be unexported")
}
