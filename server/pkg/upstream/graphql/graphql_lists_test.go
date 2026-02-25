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

// TestGraphQLUpstream_Register_ListHandling verifies that list types (and nested non-null/list types)
// are correctly handled during registration and query generation.
func TestGraphQLUpstream_Register_ListHandling(t *testing.T) {
	// Create a mock GraphQL server with a schema containing list types
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
											// Type: [ID!]!
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
									// Type: [User!]!
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
														{"name": "id"},
														{"name": "name"},
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
											// Type: [String] (Nullable list of nullable strings)
											"type": map[string]interface{}{
												"kind": "LIST",
												"ofType": map[string]interface{}{
													"name": "String",
													"kind": "SCALAR",
												},
											},
										},
									},
									// Type: Boolean
									"type": map[string]interface{}{
										"name": "Boolean",
										"kind": "SCALAR",
									},
								},
							},
						},
						// User type definition
						{
							"name": "User",
							"kind": "OBJECT",
							"fields": []map[string]interface{}{
								{
									"name": "id",
									"type": map[string]interface{}{
										"name": "ID",
										"kind": "SCALAR",
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

	// Verify 'users' tool
	usersTool, ok := toolManager.GetTool("test-service.test-service-users")
	require.True(t, ok)
	usersCallable, ok := usersTool.(*tool.CallableTool)
	require.True(t, ok)
	usersImpl, ok := usersCallable.Callable().(*Callable)
	require.True(t, ok)

	// Check the generated query string
	// Expected: query ($ids: [ID!]!) { users(ids: $ids) { id name } }
	// Note: The selection set { id name } is generated because User has fields and no explicit selection set is provided.
	assert.Contains(t, usersImpl.query, "query ($ids: [ID!]!)")
	assert.Contains(t, usersImpl.query, "users(ids: $ids)")
	assert.Contains(t, usersImpl.query, "{ id name }")

	// Verify 'addTags' tool
	tagsTool, ok := toolManager.GetTool("test-service.test-service-addTags")
	require.True(t, ok)
	tagsCallable, ok := tagsTool.(*tool.CallableTool)
	require.True(t, ok)
	tagsImpl, ok := tagsCallable.Callable().(*Callable)
	require.True(t, ok)

	// Check the generated mutation string
	// Expected: mutation ($tags: [String]) { addTags(tags: $tags) }
	// Note: Boolean is Scalar so no selection set needed (or empty).
	assert.Contains(t, tagsImpl.query, "mutation ($tags: [String])")
	assert.Contains(t, tagsImpl.query, "addTags(tags: $tags)")
}
