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

func TestGraphQLUpstream_Register_ListHandling(t *testing.T) {
	// Create a mock GraphQL server
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
															"kind": "SCALAR",
															"name": "ID",
														},
													},
												},
											},
										},
									},
									// Return type: [User!]!
									"type": map[string]interface{}{
										"kind": "NON_NULL",
										"ofType": map[string]interface{}{
											"kind": "LIST",
											"ofType": map[string]interface{}{
												"kind": "NON_NULL",
												"ofType": map[string]interface{}{
													"kind": "OBJECT",
													"name": "User",
													// IMPORTANT: Fields are here in the mocked response structure
													// The code relies on them being present here to generate default selection set.
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
											"type": map[string]interface{}{
												"kind": "LIST",
												"ofType": map[string]interface{}{
													"kind": "SCALAR",
													"name": "String",
												},
											},
										},
									},
									"type": map[string]interface{}{
										"kind": "SCALAR",
										"name": "Boolean",
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

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, nil, nil, false)
	require.NoError(t, err)

	// Verify 'users' tool
	usersTool, ok := toolManager.GetTool("test-service.test-service-users")
	require.True(t, ok)

	usersToolCallable, ok := usersTool.(*tool.CallableTool)
	require.True(t, ok)
	usersCallable, ok := usersToolCallable.Callable().(*Callable)
	require.True(t, ok)

	// Check the generated query for list argument syntax
	// Expecting: query ($ids: [ID!]!) { users(ids: $ids) ... }
	assert.Contains(t, usersCallable.query, "$ids: [ID!]!")
	assert.Contains(t, usersCallable.query, "users(ids: $ids)")

	// Check if default selection set is generated for wrapped object type
	// If the bug exists, this will likely fail because it won't find { id name }
	// It depends on whether the code unwraps the type to find fields.
	// Based on code analysis, it checks len(field.Type.Fields), which might be 0 for the wrapper.
	// But let's see.
	// We assert that it DOES contain the fields.
	// Note: The fields might be separated by space.
	assert.Contains(t, usersCallable.query, "{ id name }")

	// Verify 'addTags' tool
	addTagsTool, ok := toolManager.GetTool("test-service.test-service-addTags")
	require.True(t, ok)

	addTagsToolCallable, ok := addTagsTool.(*tool.CallableTool)
	require.True(t, ok)
	addTagsCallable, ok := addTagsToolCallable.Callable().(*Callable)
	require.True(t, ok)

	// Check the generated mutation for list argument syntax
	// Expecting: mutation ($tags: [String]) { addTags(tags: $tags) ... }
	assert.Contains(t, addTagsCallable.query, "$tags: [String]")
	assert.Contains(t, addTagsCallable.query, "addTags(tags: $tags)")
}
