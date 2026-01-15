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
)

func TestGraphQLUpstream_InputSchemaStructure(t *testing.T) {
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
									"name": "user",
									"args": []map[string]interface{}{
										{
											"name": "id",
											"type": map[string]interface{}{
												"name": "ID",
												"kind": "SCALAR",
											},
										},
										{
											"name": "active",
											"type": map[string]interface{}{
												"name": "Boolean",
												"kind": "SCALAR",
											},
										},
									},
									"type": map[string]interface{}{
										"name": "User",
										"kind": "OBJECT",
										"fields": []map[string]interface{}{
											{
												"name": "id",
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

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("schema-test-service")
	serviceConfig.SetGraphqlService(&configv1.GraphQLUpstreamService{})
	serviceConfig.GetGraphqlService().SetAddress(server.URL)

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, nil, nil, false)
	require.NoError(t, err)

	userTool, ok := toolManager.GetTool("schema-test-service.schema-test-service-user")
	require.True(t, ok)

	// Verify the input schema structure
	annotations := userTool.Tool().GetAnnotations()
	require.NotNil(t, annotations)
	inputSchema := annotations.GetInputSchema()
	require.NotNil(t, inputSchema)

	// Check if top-level "type" is "object"
	fields := inputSchema.GetFields()
	require.Contains(t, fields, "type")
	assert.Equal(t, "object", fields["type"].GetStringValue())

	// Check if "properties" exists
	require.Contains(t, fields, "properties")
	properties := fields["properties"].GetStructValue()
	require.NotNil(t, properties)

	// Check fields inside properties
	propFields := properties.GetFields()
	require.Contains(t, propFields, "id")
	require.Contains(t, propFields, "active")

	idSchema := propFields["id"].GetStructValue().GetFields()
	assert.Equal(t, "string", idSchema["type"].GetStringValue())

	activeSchema := propFields["active"].GetStructValue().GetFields()
	assert.Equal(t, "boolean", activeSchema["type"].GetStringValue())
}
