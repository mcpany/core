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
	// Create a mock GraphQL server that returns an introspection schema with List types
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
									// users(ids: [ID!]!): [User!]!
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
									// addTags(tags: [String]): Boolean
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
								{"name": "id", "type": map[string]interface{}{"name": "ID", "kind": "SCALAR"}},
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

	// Verify tool definitions
	// 1. test-service-users
	// 2. test-service-addTags
	assert.Len(t, toolDefs, 2)

	// Verify "users" tool
	usersTool, ok := toolManager.GetTool("test-service.test-service-users")
	require.True(t, ok)
	usersToolCallable, ok := usersTool.(*tool.CallableTool)
	require.True(t, ok)
	usersCallable, ok := usersToolCallable.Callable().(*Callable)
	require.True(t, ok)

	// Expected query: query ($ids: [ID!]!) { users(ids: $ids) { id } }
	// The logic builds the selection set from fields if no explicit call definition is provided.
	// For return type [User!]! (NON_NULL -> LIST -> NON_NULL -> OBJECT), it must recurse/unwrap
	// the types to find the underlying User object fields.
	// The mock schema defines User with fields [{name: "id"}].
	t.Logf("Generated Query for users: %s", usersCallable.query)
	assert.Contains(t, usersCallable.query, "users(ids: $ids) { id }")

	// Verify "addTags" tool
	addTagsTool, ok := toolManager.GetTool("test-service.test-service-addTags")
	require.True(t, ok)
	addTagsCallable, ok := addTagsTool.(*tool.CallableTool)
	require.True(t, ok)
	addTagsCallableInternal, ok := addTagsCallable.Callable().(*Callable)
	require.True(t, ok)

	// Expected query: mutation ($tags: [String]) { addTags(tags: $tags) }
	// Return type is Boolean (Scalar), so no selection set needed.
	t.Logf("Generated Query for addTags: %s", addTagsCallableInternal.query)
	assert.Contains(t, addTagsCallableInternal.query, "mutation ($tags: [String]) { addTags(tags: $tags) }")
}
