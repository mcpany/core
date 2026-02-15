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
	"google.golang.org/protobuf/types/known/structpb"
)

func TestHTTPUpstream_InputSchema_Comprehensive(t *testing.T) {
	testCases := []struct {
		name          string
		configJSON    string
		validate      func(t *testing.T, schema *structpb.Struct)
	}{
		{
			name: "InputSchema Only (No Parameters)",
			configJSON: `{
				"name": "schema-only-service",
				"http_service": {
					"address": "http://127.0.0.1",
					"tools": [{"name": "test-op", "call_id": "test-op-call"}],
					"calls": {
						"test-op-call": {
							"id": "test-op-call",
							"method": "HTTP_METHOD_POST",
							"endpoint_path": "/test",
							"input_schema": {
								"type": "object",
								"properties": {
									"foo": { "type": "string" }
								},
								"required": ["foo"]
							}
						}
					}
				}
			}`,
			validate: func(t *testing.T, schema *structpb.Struct) {
				props := schema.Fields["properties"].GetStructValue().GetFields()
				assert.Contains(t, props, "foo")
				req := schema.Fields["required"].GetListValue().GetValues()
				assert.Len(t, req, 1)
				assert.Equal(t, "foo", req[0].GetStringValue())
			},
		},
		{
			name: "Parameters Only (No InputSchema)",
			configJSON: `{
				"name": "params-only-service",
				"http_service": {
					"address": "http://127.0.0.1",
					"tools": [{"name": "test-op", "call_id": "test-op-call"}],
					"calls": {
						"test-op-call": {
							"id": "test-op-call",
							"method": "HTTP_METHOD_GET",
							"endpoint_path": "/test",
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
			}`,
			validate: func(t *testing.T, schema *structpb.Struct) {
				props := schema.Fields["properties"].GetStructValue().GetFields()
				assert.Contains(t, props, "bar")
				req := schema.Fields["required"].GetListValue().GetValues()
				assert.Len(t, req, 1)
				assert.Equal(t, "bar", req[0].GetStringValue())
			},
		},
		{
			name: "Both Defined (Merge Properties)",
			configJSON: `{
				"name": "both-defined-service",
				"http_service": {
					"address": "http://127.0.0.1",
					"tools": [{"name": "test-op", "call_id": "test-op-call"}],
					"calls": {
						"test-op-call": {
							"id": "test-op-call",
							"method": "HTTP_METHOD_POST",
							"endpoint_path": "/test",
							"input_schema": {
								"type": "object",
								"properties": {
									"body_prop": { "type": "string" }
								}
							},
							"parameters": [
								{
									"schema": {
										"name": "query_prop",
										"type": "STRING"
									}
								}
							]
						}
					}
				}
			}`,
			validate: func(t *testing.T, schema *structpb.Struct) {
				props := schema.Fields["properties"].GetStructValue().GetFields()
				assert.Contains(t, props, "body_prop")
				assert.Contains(t, props, "query_prop")
			},
		},
		{
			name: "Both Defined (InputSchema Properties Priority)",
			// If a property exists in both, InputSchema should generally NOT be overwritten by Parameters
			// if it was already present. The code does:
			// for k, v := range properties.Fields { if _, ok := existingProps[k]; !ok { existingProps[k] = v } }
			// So InputSchema properties take precedence.
			configJSON: `{
				"name": "priority-service",
				"http_service": {
					"address": "http://127.0.0.1",
					"tools": [{"name": "test-op", "call_id": "test-op-call"}],
					"calls": {
						"test-op-call": {
							"id": "test-op-call",
							"method": "HTTP_METHOD_POST",
							"endpoint_path": "/test",
							"input_schema": {
								"type": "object",
								"properties": {
									"conflict": { "type": "string", "description": "from_schema" }
								}
							},
							"parameters": [
								{
									"schema": {
										"name": "conflict",
										"type": "INTEGER",
										"description": "from_params"
									}
								}
							]
						}
					}
				}
			}`,
			validate: func(t *testing.T, schema *structpb.Struct) {
				props := schema.Fields["properties"].GetStructValue().GetFields()
				assert.Contains(t, props, "conflict")
				// Check description to see which one won
				desc := props["conflict"].GetStructValue().Fields["description"].GetStringValue()
				assert.Equal(t, "from_schema", desc)
			},
		},
		{
			name: "Both Defined (Merge Required)",
			configJSON: `{
				"name": "merge-required-service",
				"http_service": {
					"address": "http://127.0.0.1",
					"tools": [{"name": "test-op", "call_id": "test-op-call"}],
					"calls": {
						"test-op-call": {
							"id": "test-op-call",
							"method": "HTTP_METHOD_POST",
							"endpoint_path": "/test",
							"input_schema": {
								"type": "object",
								"properties": {
									"p1": { "type": "string" }
								},
								"required": ["p1"]
							},
							"parameters": [
								{
									"schema": {
										"name": "p2",
										"type": "STRING",
										"is_required": true
									}
								}
							]
						}
					}
				}
			}`,
			validate: func(t *testing.T, schema *structpb.Struct) {
				req := schema.Fields["required"].GetListValue().GetValues()
				var reqStrs []string
				for _, v := range req {
					reqStrs = append(reqStrs, v.GetStringValue())
				}
				assert.Contains(t, reqStrs, "p1")
				assert.Contains(t, reqStrs, "p2")
			},
		},
		{
			name: "Both Defined (Deduplicate Required)",
			configJSON: `{
				"name": "dedup-required-service",
				"http_service": {
					"address": "http://127.0.0.1",
					"tools": [{"name": "test-op", "call_id": "test-op-call"}],
					"calls": {
						"test-op-call": {
							"id": "test-op-call",
							"method": "HTTP_METHOD_POST",
							"endpoint_path": "/test",
							"input_schema": {
								"type": "object",
								"properties": {
									"p1": { "type": "string" }
								},
								"required": ["p1"]
							},
							"parameters": [
								{
									"schema": {
										"name": "p1",
										"type": "STRING",
										"is_required": true
									}
								}
							]
						}
					}
				}
			}`,
			validate: func(t *testing.T, schema *structpb.Struct) {
				req := schema.Fields["required"].GetListValue().GetValues()
				var reqStrs []string
				for _, v := range req {
					reqStrs = append(reqStrs, v.GetStringValue())
				}
				// Should only contain "p1" once
				count := 0
				for _, s := range reqStrs {
					if s == "p1" {
						count++
					}
				}
				assert.Equal(t, 1, count, "required list should not contain duplicates")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pm := pool.NewManager()
			tm := tool.NewManager(nil)
			upstream := NewUpstream(pm)

			serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
			require.NoError(t, protojson.Unmarshal([]byte(tc.configJSON), serviceConfig))

			serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
			assert.NoError(t, err)

			sanitizedToolName, _ := util.SanitizeToolName("test-op")
			toolID := serviceID + "." + sanitizedToolName
			registeredTool, ok := tm.GetTool(toolID)
			assert.True(t, ok)

			schema := registeredTool.Tool().GetAnnotations().GetInputSchema()
			require.NotNil(t, schema)
			tc.validate(t, schema)
		})
	}
}
