// Copyright 2026 Author(s) of MCP Any
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

func TestGraphQLUpstream_Register_ListHandling(t *testing.T) {
	// Mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"__schema": map[string]interface{}{
					"queryType":    map[string]string{"name": "Query"},
					"mutationType": map[string]string{"name": "Mutation"},
					"types": []map[string]interface{}{
						{
							"name": "Query",
							"kind": "OBJECT",
							"fields": []map[string]interface{}{
								{
									"name": "users",
									"args": []map[string]interface{}{
										{
											"name": "ids",
											"type": map[string]interface{}{
												"kind": "NON_NULL",
												"ofType": map[string]interface{}{
													"kind": "LIST",
													"ofType": map[string]interface{}{
														"kind": "NON_NULL",
														"ofType": map[string]interface{}{
															"name": "ID",
															"kind": "SCALAR",
														},
													},
												},
											},
										},
									},
									"type": map[string]interface{}{
										"kind": "NON_NULL",
										"ofType": map[string]interface{}{
											"kind": "LIST",
											"ofType": map[string]interface{}{
												"kind": "NON_NULL",
												"ofType": map[string]interface{}{
													"name": "User",
													"kind": "OBJECT",
													"fields": []map[string]interface{}{
														{
															"name": "id",
															"type": map[string]interface{}{
																"kind":   "NON_NULL",
																"ofType": map[string]interface{}{"name": "ID", "kind": "SCALAR"},
															},
														},
														{
															"name": "name",
															"type": map[string]interface{}{
																"name": "String",
																"kind": "SCALAR",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						{
							"name": "Mutation",
							"kind": "OBJECT",
							"fields": []map[string]interface{}{
								{
									"name": "addTags",
									"args": []map[string]interface{}{
										{
											"name": "tags",
											"type": map[string]interface{}{
												"kind": "LIST",
												"ofType": map[string]interface{}{
													"name": "String",
													"kind": "SCALAR",
												},
											},
										},
									},
									"type": map[string]interface{}{
										"name": "Boolean",
										"kind": "SCALAR",
									},
								},
							},
						},
						{
							"name": "User",
							"kind": "OBJECT",
							"fields": []map[string]interface{}{
								{
									"name": "id",
									"type": map[string]interface{}{
										"kind":   "NON_NULL",
										"ofType": map[string]interface{}{"name": "ID", "kind": "SCALAR"},
									},
								},
								{
									"name": "name",
									"type": map[string]interface{}{
										"name": "String",
										"kind": "SCALAR",
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

	serviceKey, toolDefs, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, nil, nil, false)
	require.NoError(t, err)

	assert.Equal(t, "test-service", serviceKey)
	assert.Len(t, toolDefs, 2)

	// Verify "users" query tool
	usersTool, ok := toolManager.GetTool("test-service.test-service-users")
	require.True(t, ok)
	usersCallable, ok := usersTool.(*tool.CallableTool)
	require.True(t, ok)
	usersQuery, ok := usersCallable.Callable().(*Callable)
	require.True(t, ok)

	// Check generated query for correct list/non-null syntax
	// Expected: query ($ids: [ID!]!) { users(ids: $ids) { id name } }
	assert.Contains(t, usersQuery.query, "query ($ids: [ID!]!)")
	assert.Contains(t, usersQuery.query, "users(ids: $ids)")
	assert.Contains(t, usersQuery.query, "{ id name }")

	// Verify "addTags" mutation tool
	addTagsTool, ok := toolManager.GetTool("test-service.test-service-addTags")
	require.True(t, ok)
	addTagsCallable, ok := addTagsTool.(*tool.CallableTool)
	require.True(t, ok)
	addTagsQuery, ok := addTagsCallable.Callable().(*Callable)
	require.True(t, ok)

	// Check generated query for correct list syntax
	// Expected: mutation ($tags: [String]) { addTags(tags: $tags) }
	assert.Contains(t, addTagsQuery.query, "mutation ($tags: [String])")
	assert.Contains(t, addTagsQuery.query, "addTags(tags: $tags)")
	// Boolean scalar doesn't have sub-selection so it should not generate braces
}
