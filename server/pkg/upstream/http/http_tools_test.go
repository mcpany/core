package http

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestHTTPUpstream_createAndRegisterHTTPTools_Coverage(t *testing.T) {
	tests := []struct {
		name          string
		configJSON    string
		expectedFqn   string
		expectedTools int
		checkSchema   func(*testing.T, *structpb.Struct)
	}{
		{
			name: "endpoint path empty but valid address",
			configJSON: ` + "`" + `{
				"name": "test-service-coverage",
				"http_service": {
					"address": "http://localhost/api",
					"tools": [{"name": "test_tool", "call_id": "test_call"}],
					"calls": {
						"test_call": {
							"id": "test_call",
							"method": "HTTP_METHOD_GET",
							"endpoint_path": ""
						}
					}
				}
			}` + "`" + `,
			expectedFqn:   "GET http://localhost/api",
			expectedTools: 1,
		},
		{
			name: "double slash root relative",
			configJSON: ` + "`" + `{
				"name": "test-service-coverage",
				"http_service": {
					"address": "http://localhost",
					"tools": [{"name": "test_tool", "call_id": "test_call"}],
					"calls": {
						"test_call": {
							"id": "test_call",
							"method": "HTTP_METHOD_GET",
							"endpoint_path": "//foo"
						}
					}
				}
			}` + "`" + `,
			expectedFqn:   "GET http://localhost//foo",
			expectedTools: 1,
		},
		{
			name: "double slash root relative with path",
			configJSON: ` + "`" + `{
				"name": "test-service-coverage",
				"http_service": {
					"address": "http://localhost/api",
					"tools": [{"name": "test_tool", "call_id": "test_call"}],
					"calls": {
						"test_call": {
							"id": "test_call",
							"method": "HTTP_METHOD_GET",
							"endpoint_path": "//foo"
						}
					}
				}
			}` + "`" + `,
			expectedFqn:   "GET http://localhost/api//foo",
			expectedTools: 1,
		},
		{
			name: "override URL queries with invalid encodings",
			configJSON: ` + "`" + `{
				"name": "test-service-coverage",
				"http_service": {
					"address": "http://localhost/api?valid=1&invalid%x=2",
					"tools": [{"name": "test_tool", "call_id": "test_call"}],
					"calls": {
						"test_call": {
							"id": "test_call",
							"method": "HTTP_METHOD_GET",
							"endpoint_path": "/test?valid=3&another%y=4"
						}
					}
				}
			}` + "`" + `,
			expectedFqn:   "GET http://localhost/api/test?valid=3&invalid%x=2&another%y=4",
			expectedTools: 1,
		},
		{
			name: "merge schemas properly with non-empty input schema",
			configJSON: ` + "`" + `{
				"name": "test-service-coverage",
				"http_service": {
					"address": "http://localhost",
					"tools": [{"name": "test_tool", "call_id": "test_call"}],
					"calls": {
						"test_call": {
							"id": "test_call",
							"method": "HTTP_METHOD_GET",
							"endpoint_path": "/schema",
							"input_schema": {
								"fields": {
									"type": { "string_value": "object" },
									"properties": {
										"struct_value": {
											"fields": {
												"existing_prop": {
													"struct_value": {
														"fields": {
															"type": { "string_value": "string" }
														}
													}
												}
											}
										}
									},
									"required": {
										"list_value": {
											"values": [
												{ "string_value": "existing_prop" },
												{ "number_value": 123 }
											]
										}
									}
								}
							},
							"parameters": {
								"fields": {
									"new_prop": {
										"struct_value": {
											"fields": {
												"type": { "string_value": "string" }
											}
										}
									}
								}
							}
						}
					}
				}
			}` + "`" + `,
			expectedFqn:   "GET http://localhost/schema",
			expectedTools: 1,
			checkSchema: func(t *testing.T, schema *structpb.Struct) {
				require.NotNil(t, schema)
				props := schema.Fields["properties"].GetStructValue()
				require.NotNil(t, props)
				require.Contains(t, props.Fields, "existing_prop")
				require.Contains(t, props.Fields, "new_prop")

				req := schema.Fields["required"].GetListValue()
				require.NotNil(t, req)

				var reqStrings []string
				for _, v := range req.Values {
					if s := v.GetStringValue(); s != "" {
						reqStrings = append(reqStrings, s)
					}
				}
				require.Contains(t, reqStrings, "existing_prop")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUpstream(nil)
			toolManager := tool.NewManager(nil)
			resourceManager := resource.NewManager()

			serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
			err := protojson.Unmarshal([]byte(tt.configJSON), serviceConfig)
			require.NoError(t, err)

			address := serviceConfig.GetHttpService().GetAddress()

			// Call the method under test directly
			discoveredTools := u.(*Upstream).createAndRegisterHTTPTools(context.Background(), "test-service-coverage", address, serviceConfig, toolManager, resourceManager, false)

			require.Len(t, discoveredTools, tt.expectedTools)
			if tt.expectedTools > 0 {
				registeredTool, ok := toolManager.GetTool("test-service-coverage.test_tool")
				require.True(t, ok)
				require.NotNil(t, registeredTool)
				require.Equal(t, tt.expectedFqn, registeredTool.Tool().GetUnderlyingMethodFqn())

				if tt.checkSchema != nil {
					tt.checkSchema(t, registeredTool.Tool().GetAnnotations().GetInputSchema())
				}
			}
		})
	}
}

func TestHTTPUpstream_createAndRegisterHTTPTools_Resources(t *testing.T) {
	u := NewUpstream(nil)
	toolManager := tool.NewManager(nil)
	resourceManager := resource.NewManager()

	configJSON := ` + "`" + `{
		"name": "test-service-resources",
		"http_service": {
			"address": "http://localhost",
			"tools": [
				{
					"name": "test_tool",
					"call_id": "test_call"
				}
			],
			"calls": {
				"test_call": {
					"id": "test_call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/call"
				}
			},
			"resources": [
				{
					"name": "disabled_res",
					"disable": true
				},
				{
					"name": "non_exported_res",
					"disable": false
				},
				{
					"name": "dynamic_res",
					"disable": false,
					"dynamic": {
						"http_call": {
							"id": "test_call"
						}
					}
				},
				{
					"name": "dynamic_res_missing_call",
					"disable": false,
					"dynamic": {
						"http_call": {
							"id": "missing_call"
						}
					}
				},
				{
					"name": "static_res",
					"disable": false,
					"uri": "http://localhost/static"
				}
			]
		},
		"resource_export_policy": {
			"rules": [
				{
					"name_regex": "^non_exported_res$",
					"action": "UNEXPORT"
				}
			]
		}
	}` + "`" + `
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	u.(*Upstream).createAndRegisterHTTPTools(context.Background(), "test-service-resources", "http://localhost", serviceConfig, toolManager, resourceManager, false)

	// Check resources
	_, ok := resourceManager.GetResource("disabled_res")
	require.False(t, ok)

	_, ok = resourceManager.GetResource("non_exported_res")
	require.False(t, ok)

	_, ok = resourceManager.GetResource("http://localhost/static")
	require.True(t, ok)
}
