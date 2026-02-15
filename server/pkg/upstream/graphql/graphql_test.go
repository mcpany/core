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

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		GraphqlService: configv1.GraphQLUpstreamService_builder{
			Address: proto.String(server.URL),
		}.Build(),
	}.Build()

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

	userToolCallable, ok := userTool.(*tool.CallableTool)
	require.True(t, ok)
	callable, ok := userToolCallable.Callable().(*Callable)
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

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		GraphqlService: configv1.GraphQLUpstreamService_builder{
			Address: proto.String(server.URL),
			Calls: map[string]*configv1.GraphQLCallDefinition{
				"user": configv1.GraphQLCallDefinition_builder{
					SelectionSet: proto.String("{ id }"),
				}.Build(),
			},
		}.Build(),
	}.Build()

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, nil, nil, false)
	require.NoError(t, err)

	userTool, ok := toolManager.GetTool("test-service.test-service-user")
	require.True(t, ok)

	userToolCallable, ok := userTool.(*tool.CallableTool)
	require.True(t, ok)
	callable, ok := userToolCallable.Callable().(*Callable)
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

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		GraphqlService: configv1.GraphQLUpstreamService_builder{
			Address: proto.String(server.URL),
		}.Build(),
		UpstreamAuth: configv1.Authentication_builder{
			ApiKey: configv1.APIKeyAuth_builder{
				ParamName: proto.String("X-API-Key"),
				Value: configv1.SecretValue_builder{
					PlainText: proto.String("test-api-key"),
				}.Build(),
			}.Build(),
		}.Build(),
	}.Build()

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

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		GraphqlService: configv1.GraphQLUpstreamService_builder{
			Address: proto.String(server.URL),
		}.Build(),
		UpstreamAuth: configv1.Authentication_builder{
			ApiKey: configv1.APIKeyAuth_builder{
				ParamName: proto.String("X-API-Key"),
				Value: configv1.SecretValue_builder{
					PlainText: proto.String("test-api-key"),
				}.Build(),
			}.Build(),
		}.Build(),
	}.Build()

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

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		GraphqlService: configv1.GraphQLUpstreamService_builder{
			Address: proto.String(server.URL),
		}.Build(),
		UpstreamAuth: configv1.Authentication_builder{
			ApiKey: configv1.APIKeyAuth_builder{
				ParamName: proto.String("X-API-Key"),
				Value: configv1.SecretValue_builder{
					PlainText: proto.String("test-api-key"),
				}.Build(),
			}.Build(),
		}.Build(),
	}.Build()

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

	toolDef := configv1.ToolDefinition_builder{
		Name: proto.String("test-service-user"),
	}.Build()
	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		GraphqlService: configv1.GraphQLUpstreamService_builder{
			Address: proto.String(server.URL),
		}.Build(),
	}.Build()

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

	toolDef := configv1.ToolDefinition_builder{
		Name: proto.String("test-service-createUser"),
	}.Build()
	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		GraphqlService: configv1.GraphQLUpstreamService_builder{
			Address: proto.String(server.URL),
		}.Build(),
	}.Build()

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

func TestGraphQLUpstream_Register_InvalidAddress(t *testing.T) {
	upstream := NewGraphQLUpstream()
	toolManager := tool.NewManager(nil)

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service-invalid"),
		GraphqlService: configv1.GraphQLUpstreamService_builder{
			Address: proto.String("file:///etc/passwd"),
		}.Build(),
	}.Build()

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, nil, nil, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid graphql service address scheme")
}
