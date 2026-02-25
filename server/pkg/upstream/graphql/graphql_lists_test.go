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
	// Create a mock GraphQL server that returns a schema with LIST types
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

	// Verify 'users' tool query construction
	usersTool, ok := toolManager.GetTool("test-service.test-service-users")
	require.True(t, ok)
	usersCallable, ok := usersTool.(*tool.CallableTool)
	require.True(t, ok)
	callable, ok := usersCallable.Callable().(*Callable)
	require.True(t, ok)

	// users(ids: [ID!]!)
	// Structure for `ids` is NON_NULL -> LIST -> NON_NULL -> ID.
	// So it is `[ID!]!`.
	// The query variable should be `$ids: [ID!]!`.
	assert.Contains(t, callable.query, "$ids: [ID!]!")
	assert.Contains(t, callable.query, "users(ids: $ids)")

	// Verify 'addTags' tool query construction
	tagsTool, ok := toolManager.GetTool("test-service.test-service-addTags")
	require.True(t, ok)
	tagsCallable, ok := tagsTool.(*tool.CallableTool)
	require.True(t, ok)
	callableTags, ok := tagsCallable.Callable().(*Callable)
	require.True(t, ok)

	// addTags(tags: [String])
	// Structure: LIST -> String.
	// Query variable: `$tags: [String]`.
	assert.Contains(t, callableTags.query, "$tags: [String]")
	assert.Contains(t, callableTags.query, "addTags(tags: $tags)")
}
