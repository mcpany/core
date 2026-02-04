// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package graphql

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestGraphQLUpstream_Register_ListArgumentBug(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	// Create a mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"__schema": map[string]interface{}{
					"queryType":    map[string]string{"name": "Query"},
					"mutationType": nil,
					"types": []map[string]interface{}{
						{
							"name": "Query",
							"kind": "OBJECT",
							"fields": []map[string]interface{}{
								{
									"name": "search",
									"args": []map[string]interface{}{
										{
											"name": "tags",
											"type": map[string]interface{}{
												"kind": "LIST",
												"name": nil,
												"ofType": map[string]interface{}{
													"kind": "SCALAR",
													"name": "String",
												},
											},
										},
									},
									"type": map[string]interface{}{
										"name": "Result",
										"kind": "OBJECT",
										"fields": []map[string]interface{}{
											{
												"name": "count",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	upstream := NewGraphQLUpstream()
	toolManager := tool.NewManager(nil)

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		GraphqlService: configv1.GraphQLUpstreamService_builder{
			Address: proto.String(server.URL),
		}.Build(),
	}.Build()

	_, toolDefs, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, nil, nil, false)
	require.NoError(t, err)

	assert.Len(t, toolDefs, 1)
	toolDef := toolDefs[0]
	assert.Equal(t, "test-service-search", toolDef.GetName())

	// Inspect the registered tool in toolManager to get the InputSchema
	registeredTool, ok := toolManager.GetTool("test-service.test-service-search")
	require.True(t, ok)

	callableTool, ok := registeredTool.(*tool.CallableTool)
	require.True(t, ok)

	inputSchema := callableTool.Tool().GetAnnotations().GetInputSchema()

	tagsField, ok := inputSchema.Fields["tags"]
	require.True(t, ok, "tags field should exist")

	tagsStruct := tagsField.GetStructValue()
	require.NotNil(t, tagsStruct)

	typeField := tagsStruct.Fields["type"]
	require.NotNil(t, typeField)

	// This assertion verifies the bug fix
	assert.Equal(t, "array", typeField.GetStringValue(), "Expected 'tags' argument to be of type 'array'")

	itemsField := tagsStruct.Fields["items"]
	// If it was correctly identified as array, it should have items
	if typeField.GetStringValue() == "array" {
		require.NotNil(t, itemsField)
		itemsStruct := itemsField.GetStructValue()
		itemsType := itemsStruct.Fields["type"]
		assert.Equal(t, "string", itemsType.GetStringValue())
	}
}
