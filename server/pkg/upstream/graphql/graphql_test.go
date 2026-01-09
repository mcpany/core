// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package graphql

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/machinebox/graphql"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestGraphQLUpstream_Register(t *testing.T) {
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
									"name": "hello",
									"args": []map[string]interface{}{},
									"type": map[string]interface{}{
										"name": "String",
										"kind": "SCALAR",
									},
								},
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
									},
									"type": map[string]interface{}{
										"name": "User",
										"kind": "OBJECT",
										"fields": []map[string]interface{}{
											{
												"name": "id",
											},
											{
												"name": "name",
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
									"name": "createUser",
									"args": []map[string]interface{}{
										{
											"name": "name",
											"type": map[string]interface{}{
												"name":   "String",
												"kind":   "NON_NULL",
												"ofType": map[string]interface{}{"name": "String", "kind": "SCALAR"},
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
											{
												"name": "name",
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
	serviceConfig.SetName("test-service")
	serviceConfig.SetGraphqlService(&configv1.GraphQLUpstreamService{})
	serviceConfig.GetGraphqlService().SetAddress(server.URL)

	serviceKey, toolDefs, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, nil, nil, false)
	require.NoError(t, err)

	assert.Equal(t, "test-service", serviceKey)
	assert.Len(t, toolDefs, 3)
	assert.Equal(t, "test-service-hello", toolDefs[0].GetName())
	assert.Equal(t, "test-service-user", toolDefs[1].GetName())
	assert.Equal(t, "test-service-createUser", toolDefs[2].GetName())

	_, ok := toolManager.GetTool("test-service.test-service-hello")
	assert.True(t, ok)
	_, ok = toolManager.GetTool("test-service.test-service-user")
	assert.True(t, ok)
	_, ok = toolManager.GetTool("test-service.test-service-createUser")
	assert.True(t, ok)

	userTool, ok := toolManager.GetTool("test-service.test-service-user")
	require.True(t, ok)

	callableTool, ok := userTool.(*tool.CallableTool)
	require.True(t, ok)
	callable, ok := callableTool.Callable().(*Callable)
	require.True(t, ok)

	assert.Contains(t, callable.query, "user(id: $id) { id name }")
}

func TestGraphQLUpstream_RegisterWithSelectionSet(t *testing.T) {
	// Create a mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"__schema": map[string]interface{}{
					"queryType": map[string]string{"name": "Query"},
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
									},
									"type": map[string]interface{}{
										"name": "User",
										"kind": "OBJECT",
										"fields": []map[string]interface{}{
											{
												"name": "id",
											},
											{
												"name": "name",
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
	serviceConfig.SetName("test-service")
	serviceConfig.SetGraphqlService(&configv1.GraphQLUpstreamService{})
	serviceConfig.GetGraphqlService().SetAddress(server.URL)
	selectionSet := "{ id }"
	calls := make(map[string]*configv1.GraphQLCallDefinition)
	calls["user"] = &configv1.GraphQLCallDefinition{}
	calls["user"].SetSelectionSet(selectionSet)
	serviceConfig.GetGraphqlService().SetCalls(calls)

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, nil, nil, false)
	require.NoError(t, err)

	userTool, ok := toolManager.GetTool("test-service.test-service-user")
	require.True(t, ok)

	callableTool, ok := userTool.(*tool.CallableTool)
	require.True(t, ok)
	callable, ok := callableTool.Callable().(*Callable)
	require.True(t, ok)

	assert.Contains(t, callable.query, "user(id: $id) { id }")
}

func TestGraphQLUpstream_RegisterWithAPIKeyAuth(t *testing.T) {
	// Create a mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-api-key", r.Header.Get("X-API-Key"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var reqBody map[string]interface{}
		err = json.Unmarshal(body, &reqBody)
		require.NoError(t, err)

		var response map[string]interface{}
		if strings.Contains(reqBody["query"].(string), "IntrospectionQuery") {
			response = map[string]interface{}{
				"data": map[string]interface{}{
					"__schema": map[string]interface{}{
						"queryType": map[string]string{"name": "Query"},
						"types": []map[string]interface{}{
							{
								"name": "Query",
								"kind": "OBJECT",
								"fields": []map[string]interface{}{
									{
										"name": "hello",
										"args": []map[string]interface{}{},
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
		} else {
			response = map[string]interface{}{
				"data": map[string]interface{}{
					"hello": "world",
				},
			}
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	upstream := NewGraphQLUpstream()
	toolManager := tool.NewManager(nil)

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-service")
	serviceConfig.SetGraphqlService(&configv1.GraphQLUpstreamService{})
	serviceConfig.GetGraphqlService().SetAddress(server.URL)
	apiKeyAuth := &configv1.APIKeyAuth{
		ParamName: proto.String("X-API-Key"),
		Value: &configv1.SecretValue{
			Value: &configv1.SecretValue_PlainText{PlainText: "test-api-key"},
		},
	}
	authConfig := &configv1.Authentication{
		AuthMethod: &configv1.Authentication_ApiKey{
			ApiKey: apiKeyAuth,
		},
	}
	serviceConfig.SetUpstreamAuth(authConfig)

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, nil, nil, false)
	require.NoError(t, err)

	helloTool, ok := toolManager.GetTool("test-service.test-service-hello")
	require.True(t, ok)

	req := &tool.ExecutionRequest{
		Arguments: map[string]interface{}{},
	}

	resp, err := helloTool.Execute(context.Background(), req)
	require.NoError(t, err)

	respMap, ok := resp.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "world", respMap["hello"])
}

func TestGraphQLUpstream_RegisterWithAPIKeyAuth_IntrospectionFails(t *testing.T) {
	// Create a mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	upstream := NewGraphQLUpstream()
	toolManager := tool.NewManager(nil)

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-service")
	serviceConfig.SetGraphqlService(&configv1.GraphQLUpstreamService{})
	serviceConfig.GetGraphqlService().SetAddress(server.URL)
	apiKeyAuth := &configv1.APIKeyAuth{
		ParamName: proto.String("X-API-Key"),
		Value: &configv1.SecretValue{
			Value: &configv1.SecretValue_PlainText{PlainText: "test-api-key"},
		},
	}
	authConfig := &configv1.Authentication{
		AuthMethod: &configv1.Authentication_ApiKey{
			ApiKey: apiKeyAuth,
		},
	}
	serviceConfig.SetUpstreamAuth(authConfig)

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, nil, nil, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to run introspection query")
}

func TestGraphQLUpstream_RegisterWithAPIKeyAuth_ToolCallFails(t *testing.T) {
	// Create a mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-api-key", r.Header.Get("X-API-Key"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var reqBody map[string]interface{}
		err = json.Unmarshal(body, &reqBody)
		require.NoError(t, err)

		if strings.Contains(reqBody["query"].(string), "IntrospectionQuery") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"__schema": map[string]interface{}{
						"queryType": map[string]string{"name": "Query"},
						"types": []map[string]interface{}{
							{
								"name": "Query",
								"kind": "OBJECT",
								"fields": []map[string]interface{}{
									{
										"name": "hello",
										"args": []map[string]interface{}{},
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
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer server.Close()

	upstream := NewGraphQLUpstream()
	toolManager := tool.NewManager(nil)

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-service")
	serviceConfig.SetGraphqlService(&configv1.GraphQLUpstreamService{})
	serviceConfig.GetGraphqlService().SetAddress(server.URL)
	apiKeyAuth := &configv1.APIKeyAuth{
		ParamName: proto.String("X-API-Key"),
		Value: &configv1.SecretValue{
			Value: &configv1.SecretValue_PlainText{PlainText: "test-api-key"},
		},
	}
	authConfig := &configv1.Authentication{
		AuthMethod: &configv1.Authentication_ApiKey{
			ApiKey: apiKeyAuth,
		},
	}
	serviceConfig.SetUpstreamAuth(authConfig)

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, nil, nil, false)
	require.NoError(t, err)

	helloTool, ok := toolManager.GetTool("test-service.test-service-hello")
	require.True(t, ok)

	req := &tool.ExecutionRequest{
		Arguments: map[string]interface{}{},
	}

	_, err = helloTool.Execute(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to run graphql query")
}

func TestGraphQLTool_ExecuteQuery(t *testing.T) {
	// Create a mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"id":   "1",
					"name": "test",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	toolDef := &configv1.ToolDefinition{}
	toolDef.SetName("test-service-user")
	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-service")
	serviceConfig.SetGraphqlService(&configv1.GraphQLUpstreamService{})
	serviceConfig.GetGraphqlService().SetAddress(server.URL)

	callable := &Callable{
		client: graphql.NewClient(server.URL),
		query:  `query ($id: ID) { user(id: $id) { id name } }`,
	}
	graphqlTool, err := tool.NewCallableTool(toolDef, serviceConfig, callable, nil, nil)
	require.NoError(t, err)

	req := &tool.ExecutionRequest{
		Arguments: map[string]interface{}{
			"id": "1",
		},
	}

	resp, err := graphqlTool.Execute(context.Background(), req)
	require.NoError(t, err)

	respMap, ok := resp.(map[string]interface{})
	require.True(t, ok)

	user, ok := respMap["user"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "1", user["id"])
	assert.Equal(t, "test", user["name"])
}

func TestGraphQLTool_ExecuteMutation(t *testing.T) {
	// Create a mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"createUser": map[string]interface{}{
					"id":   "2",
					"name": "new-user",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	toolDef := &configv1.ToolDefinition{}
	toolDef.SetName("test-service-createUser")
	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-service")
	serviceConfig.SetGraphqlService(&configv1.GraphQLUpstreamService{})
	serviceConfig.GetGraphqlService().SetAddress(server.URL)

	callable := &Callable{
		client: graphql.NewClient(server.URL),
		query:  `mutation ($name: String!) { createUser(name: $name) { id name } }`,
	}
	graphqlTool, err := tool.NewCallableTool(toolDef, serviceConfig, callable, nil, nil)
	require.NoError(t, err)

	req := &tool.ExecutionRequest{
		Arguments: map[string]interface{}{
			"name": "new-user",
		},
	}

	resp, err := graphqlTool.Execute(context.Background(), req)
	require.NoError(t, err)

	respMap, ok := resp.(map[string]interface{})
	require.True(t, ok)

	user, ok := respMap["createUser"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "2", user["id"])
	assert.Equal(t, "new-user", user["name"])
}
