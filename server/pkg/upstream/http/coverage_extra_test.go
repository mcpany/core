// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestCoverageExtra_Register_Reload(t *testing.T) {
	pm := pool.NewManager()
	tm := newMockToolManager()
	upstream := NewUpstream(pm)
	defer upstream.Shutdown(context.Background())

	configJSON := `{
		"name": "reload-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "op1", "call_id": "c1"}],
			"calls": {
				"c1": {"id": "c1", "method": "HTTP_METHOD_GET"}
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	// First registration
	serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	// Verify tools added
	assert.Len(t, tm.addedTools, 1)

	// Second registration (Reload)
	// mockToolManager.ClearToolsForService logic is simple: it removes tools matching serviceID.
	// We want to ensure it is called.

	// Reset addedTools to verify clear works (or at least that Register calls it)
	// But Register calls ClearToolsForService, then adds tools.
	// So if we reload with SAME config, we should end up with 1 tool (old one cleared, new one added).
	// If Clear was NOT called, we might end up with duplicates or mock might not support duplicates.

	_, _, _, err = upstream.Register(context.Background(), serviceConfig, tm, nil, nil, true)
	require.NoError(t, err)

	// Check that we still have 1 tool (effectively verifying clean up happened, assuming mock works correctly)
	// Actually, our mock implementation of ClearToolsForService removes tools.
	assert.Len(t, tm.addedTools, 1)
	assert.Equal(t, serviceID, tm.addedTools[0].Tool().GetServiceId())
}

func TestCoverageExtra_InvalidScheme(t *testing.T) {
	pm := pool.NewManager()
	tm := newMockToolManager()
	upstream := NewUpstream(pm)
	defer upstream.Shutdown(context.Background())

	configJSON := `{
		"name": "ws-scheme-service",
		"http_service": {
			"address": "ws://127.0.0.1",
			"tools": [{"name": "op1", "call_id": "c1"}],
			"calls": {
				"c1": {"id": "c1", "method": "HTTP_METHOD_GET"}
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid http service address scheme")
	assert.Contains(t, err.Error(), "ws")
}

func TestCoverageExtra_InputSchema_ComplexMerge(t *testing.T) {
	pm := pool.NewManager()
	tm := newMockToolManager()
	upstream := NewUpstream(pm)
	defer upstream.Shutdown(context.Background())

	// Schema has required ["a"].
	// Parameters has "b" (required).
	// Merged schema should have required ["a", "b"].

	configJSON := `{
		"name": "merge-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "op1", "call_id": "c1"}],
			"calls": {
				"c1": {
					"id": "c1",
					"method": "HTTP_METHOD_GET",
					"input_schema": {
						"type": "object",
						"properties": {
							"a": {"type": "string"}
						},
						"required": ["a"]
					},
					"parameters": [
						{
							"schema": {
								"name": "b",
								"type": "STRING",
								"is_required": true
							}
						}
					]
				}
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	require.Len(t, tm.addedTools, 1)
	toolDef := tm.addedTools[0].Tool()
	schema := toolDef.GetAnnotations().GetInputSchema()

	// Check required
	reqVal := schema.Fields["required"].GetListValue()
	require.NotNil(t, reqVal)

	requiredSet := make(map[string]bool)
	for _, v := range reqVal.Values {
		requiredSet[v.GetStringValue()] = true
	}

	assert.True(t, requiredSet["a"], "a should be required")
	assert.True(t, requiredSet["b"], "b should be required")

	// Check properties
	props := schema.Fields["properties"].GetStructValue()
	require.NotNil(t, props)
	assert.Contains(t, props.Fields, "a")
	assert.Contains(t, props.Fields, "b")
}

func TestCoverageExtra_URLParsingError_Address(t *testing.T) {
	pm := pool.NewManager()
	tm := newMockToolManager()
	upstream := NewUpstream(pm)
	defer upstream.Shutdown(context.Background())

	// Control char in address
	configJSON := `{
		"name": "bad-address-service",
		"http_service": {
			"address": "http://127.0.0.1/\u007f",
			"tools": [{"name": "op1", "call_id": "c1"}],
			"calls": {
				"c1": {"id": "c1", "method": "HTTP_METHOD_GET"}
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid http service address")
}

func TestCoverageExtra_DisabledComponents(t *testing.T) {
	pm := pool.NewManager()
	tm := newMockToolManager()
	rm := resource.NewManager()
	prm := prompt.NewManager()
	upstream := NewUpstream(pm)
	defer upstream.Shutdown(context.Background())

	configJSON := `{
		"name": "disabled-comps-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [
				{"name": "t1", "call_id": "c1", "disable": true},
				{"name": "t2", "call_id": "c1", "disable": false}
			],
			"calls": {
				"c1": {"id": "c1", "method": "HTTP_METHOD_GET"}
			},
			"resources": [
				{"name": "r1", "uri": "http://r1", "disable": true},
				{"name": "r2", "uri": "http://r2", "disable": false}
			],
			"prompts": [
				{"name": "p1", "disable": true},
				{"name": "p2", "disable": false}
			]
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, prm, rm, false)
	require.NoError(t, err)

	// Tools
	assert.Len(t, discoveredTools, 1)
	assert.Equal(t, "t2", discoveredTools[0].GetName())

	// Resources
	_, ok := rm.GetResource("http://r1")
	assert.False(t, ok)
	_, ok = rm.GetResource("http://r2")
	assert.True(t, ok)

	// Prompts
	// We need to calculate prompt ID. Prompts are added by NewTemplatedPrompt(promptDef, serviceID)
	// TemplatedPrompt name is just the name from def.
	// But PromptManager keys them by ID?
	// promptManager.AddPrompt uses p.Prompt().Name usually?
	// Let's check prompt.Manager interface or implementation.
	// Standard Manager usually keys by name if unique, or name is the ID.
	// In Register: promptManager.AddPrompt(newPrompt)
	// let's assume get works by prompt name + service ID prefix?
	// Or maybe just prompt name?
	// NewTemplatedPrompt stores serviceID.

	// Let's check prompt names in prm.
	// We don't have direct list access to prm (it's interface in Register, but we passed concrete).
	// But we can check if we can list them or get them.

	// Assuming sanitized service name + prompt name
	sanitizedServiceName, _ := util.SanitizeServiceName("disabled-comps-service")

	// Check p1 (disabled)
	sanitizedP1, _ := util.SanitizeToolName("p1")
	p1ID := sanitizedServiceName + "." + sanitizedP1
	_, ok = prm.GetPrompt(p1ID)
	assert.False(t, ok)

	// Check p2 (enabled)
	sanitizedP2, _ := util.SanitizeToolName("p2")
	p2ID := sanitizedServiceName + "." + sanitizedP2
	_, ok = prm.GetPrompt(p2ID)
	assert.True(t, ok)
}

func TestCoverageExtra_InputSchema_InvalidRequiredType(t *testing.T) {
	pm := pool.NewManager()
	tm := newMockToolManager()
	upstream := NewUpstream(pm)
	defer upstream.Shutdown(context.Background())

	// "required" is not a list (e.g. string "foo").
	// Should be overwritten by parameters required.

	configJSON := `{
		"name": "invalid-req-type-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "op1", "call_id": "c1"}],
			"calls": {
				"c1": {
					"id": "c1",
					"method": "HTTP_METHOD_GET",
					"input_schema": {
						"type": "object",
						"required": "foo"
					},
					"parameters": [
						{
							"schema": {
								"name": "bar",
								"type": "STRING",
								"is_required": true
							}
						}
					]
				}
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	require.Len(t, tm.addedTools, 1)
	toolDef := tm.addedTools[0].Tool()
	schema := toolDef.GetAnnotations().GetInputSchema()

	reqVal := schema.Fields["required"].GetListValue()
	require.NotNil(t, reqVal)
	require.Len(t, reqVal.Values, 1)
	assert.Equal(t, "bar", reqVal.Values[0].GetStringValue())
}

func TestCoverageExtra_URL_WithUser(t *testing.T) {
	pm := pool.NewManager()
	tm := newMockToolManager()
	upstream := NewUpstream(pm)
	defer upstream.Shutdown(context.Background())

	// endpoint_path with //user:pass@host/path

	configJSON := `{
		"name": "url-user-service",
		"http_service": {
			"address": "http://127.0.0.1/base",
			"tools": [{"name": "op1", "call_id": "c1"}],
			"calls": {
				"c1": {
					"id": "c1",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "//user:pass@example.com/path"
				}
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	require.Len(t, tm.addedTools, 1)
	fqn := tm.addedTools[0].Tool().GetUnderlyingMethodFqn()

	// Should be appended to base?
	// Logic: //user:pass@host/path -> path starts with //.
	// Becomes path relative to base?
	// Wait, if it's scheme-relative, it should be treated as relative path to base host?
	// The code:
	// if endpointURL.Host != "" {
	//    prefix += ...
	//    endpointURL.Path = prefix + endpointURL.Path
	//    endpointURL.Host = ""
	// }
	// So path becomes "//user:pass@example.com/path"
	// Then merged with base "http://127.0.0.1/base".
	// Result: "http://127.0.0.1/base//user:pass@example.com/path"
	// Verify this.

	assert.Contains(t, fqn, "http://127.0.0.1/base//user:pass@example.com/path")
}

func TestCoverageExtra_ExportPolicy_ResourcesAndPrompts(t *testing.T) {
	pm := pool.NewManager()
	tm := newMockToolManager()
	rm := resource.NewManager()
	prm := prompt.NewManager()
	upstream := NewUpstream(pm)
	defer upstream.Shutdown(context.Background())

	configJSON := `{
		"name": "export-policy-extra",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "op1", "call_id": "c1"}],
			"calls": {
				"c1": {"id": "c1", "method": "HTTP_METHOD_GET"}
			},
			"resources": [
				{"name": "public-res", "uri": "http://r1"},
				{"name": "private-res", "uri": "http://r2"}
			],
			"prompts": [
				{"name": "public-prompt"},
				{"name": "private-prompt"}
			]
		},
		"resource_export_policy": {
			"default_action": "EXPORT",
			"rules": [{"name_regex": "^private-.*", "action": "UNEXPORT"}]
		},
		"prompt_export_policy": {
			"default_action": "EXPORT",
			"rules": [{"name_regex": "^private-.*", "action": "UNEXPORT"}]
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, prm, rm, false)
	require.NoError(t, err)

	// Resources
	_, ok := rm.GetResource("http://r1")
	assert.True(t, ok, "public-res should be exported")
	_, ok = rm.GetResource("http://r2")
	assert.False(t, ok, "private-res should not be exported")

	// Prompts
	sanitizedServiceName, _ := util.SanitizeServiceName("export-policy-extra")

	sanitizedPublic, _ := util.SanitizeToolName("public-prompt")
	_, ok = prm.GetPrompt(sanitizedServiceName + "." + sanitizedPublic)
	assert.True(t, ok, "public-prompt should be exported")

	sanitizedPrivate, _ := util.SanitizeToolName("private-prompt")
	_, ok = prm.GetPrompt(sanitizedServiceName + "." + sanitizedPrivate)
	assert.False(t, ok, "private-prompt should not be exported")
}

func TestCoverageExtra_FragmentInheritance(t *testing.T) {
	pm := pool.NewManager()
	tm := newMockToolManager()
	upstream := NewUpstream(pm)
	defer upstream.Shutdown(context.Background())

	// Base has fragment #base. Endpoint has no fragment -> Result #base.
	// Base has fragment #base. Endpoint has fragment #end -> Result #end.

	configJSON := `{
		"name": "fragment-service",
		"http_service": {
			"address": "http://127.0.0.1#base",
			"tools": [
				{"name": "op1", "call_id": "c1"},
				{"name": "op2", "call_id": "c2"}
			],
			"calls": {
				"c1": {"id": "c1", "method": "HTTP_METHOD_GET"},
				"c2": {"id": "c2", "method": "HTTP_METHOD_GET", "endpoint_path": "/foo#end"}
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	require.Len(t, tm.addedTools, 2)

	// op1: check fragment #base
	// op2: check fragment #end

	for _, toolItem := range tm.addedTools {
		name := toolItem.Tool().GetName()
		fqn := toolItem.Tool().GetUnderlyingMethodFqn()
		if name == "op1" {
			assert.Contains(t, fqn, "#base")
			assert.NotContains(t, fqn, "#end")
		} else if name == "op2" {
			assert.Contains(t, fqn, "#end")
		}
	}
}

func TestCoverageExtra_NoInputSchema_NoParams(t *testing.T) {
	pm := pool.NewManager()
	tm := newMockToolManager()
	upstream := NewUpstream(pm)
	defer upstream.Shutdown(context.Background())

	configJSON := `{
		"name": "no-schema-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "op1", "call_id": "c1"}],
			"calls": {
				"c1": {"id": "c1", "method": "HTTP_METHOD_GET"}
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	require.Len(t, tm.addedTools, 1)
	schema := tm.addedTools[0].Tool().GetAnnotations().GetInputSchema()
	require.NotNil(t, schema)
	// Should be type object, empty properties
	assert.Equal(t, "object", schema.Fields["type"].GetStringValue())
	props := schema.Fields["properties"].GetStructValue()
	assert.Empty(t, props.Fields)
}

func TestCoverageExtra_NoInputSchema_WithParams(t *testing.T) {
	pm := pool.NewManager()
	tm := newMockToolManager()
	upstream := NewUpstream(pm)
	defer upstream.Shutdown(context.Background())

	configJSON := `{
		"name": "params-only-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "op1", "call_id": "c1"}],
			"calls": {
				"c1": {
					"id": "c1", "method": "HTTP_METHOD_GET",
					"parameters": [
						{"schema": {"name": "p1", "type": "STRING", "is_required": true}}
					]
				}
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	require.Len(t, tm.addedTools, 1)
	schema := tm.addedTools[0].Tool().GetAnnotations().GetInputSchema()
	require.NotNil(t, schema)

	props := schema.Fields["properties"].GetStructValue()
	assert.Contains(t, props.Fields, "p1")

	req := schema.Fields["required"].GetListValue()
	require.Len(t, req.Values, 1)
	assert.Equal(t, "p1", req.Values[0].GetStringValue())
}

func TestCoverageExtra_CheckHealth_Success(t *testing.T) {
	pm := pool.NewManager()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	configJSON := `{
		"name": "health-success-service",
		"http_service": {
			"address": "` + server.URL + `",
			"tools": [{"name": "op1", "call_id": "c1"}],
			"calls": {
				"c1": {"id": "c1", "method": "HTTP_METHOD_GET"}
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	upstream := NewUpstream(pm)
	defer upstream.Shutdown(context.Background())

	// Register
	_, _, _, err := upstream.Register(context.Background(), serviceConfig, newMockToolManager(), nil, nil, false)
	require.NoError(t, err)

	// Check Health
	hc, ok := upstream.(interface{ CheckHealth(context.Context) error })
	require.True(t, ok)

	err = hc.CheckHealth(context.Background())
	assert.NoError(t, err)
}
