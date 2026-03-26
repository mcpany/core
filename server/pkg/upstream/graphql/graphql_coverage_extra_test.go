// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package graphql

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
)

// TestGraphQLUpstream_Register_EmptyAddress ...
// Summary: TestGraphQLUpstream_Register_EmptyAddress
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	u := NewGraphQLUpstream()
	tm := tool.NewManager(nil)

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		GraphqlService: configv1.GraphQLUpstreamService_builder{
			Address: proto.String(""),
		}.Build(),
	}.Build()

	_, _, _, err := u.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "graphql service address is required")
}

// TestGraphQLUpstream_Register_InvalidURL ...
// Summary: TestGraphQLUpstream_Register_InvalidURL
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	u := NewGraphQLUpstream()
	tm := tool.NewManager(nil)

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		GraphqlService: configv1.GraphQLUpstreamService_builder{
			Address: proto.String("http://[::1]:namedport"),
		}.Build(),
	}.Build()

	_, _, _, err := u.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid graphql service address")
}

// TestGraphQLUpstream_Register_AuthCreationError ...
// Summary: TestGraphQLUpstream_Register_AuthCreationError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	u := NewGraphQLUpstream()
	tm := tool.NewManager(nil)

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		GraphqlService: configv1.GraphQLUpstreamService_builder{
			Address: proto.String("http://example.com"),
		}.Build(),
		UpstreamAuth: configv1.Authentication_builder{
			ApiKey: configv1.APIKeyAuth_builder{
				Value: configv1.SecretValue_builder{
					PlainText: proto.String("key"),
				}.Build(),
				// Missing ParamName
			}.Build(),
		}.Build(),
	}.Build()

	_, _, _, err := u.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create upstream authenticator")
}

// TestFormatGraphQLType ...
// Summary: TestFormatGraphQLType
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	tests := []struct {
		name     string
		input    *graphQLType
		expected string
	}{
		{
			name:     "Nil",
			input:    nil,
			expected: "",
		},
		{
			name: "Scalar",
			input: &graphQLType{
				Kind: "SCALAR",
				Name: proto.String("String"),
			},
			expected: "String",
		},
		{
			name: "NonNull",
			input: &graphQLType{
				Kind: "NON_NULL",
				OfType: &graphQLType{
					Kind: "SCALAR",
					Name: proto.String("String"),
				},
			},
			expected: "String!",
		},
		{
			name: "List",
			input: &graphQLType{
				Kind: "LIST",
				OfType: &graphQLType{
					Kind: "SCALAR",
					Name: proto.String("Int"),
				},
			},
			expected: "[Int]",
		},
		{
			name: "ListNonNull",
			input: &graphQLType{
				Kind: "LIST",
				OfType: &graphQLType{
					Kind: "NON_NULL",
					OfType: &graphQLType{
						Kind: "SCALAR",
						Name: proto.String("Int"),
					},
				},
			},
			expected: "[Int!]",
		},
		{
			name: "NamelessScalar",
			input: &graphQLType{
				Kind: "SCALAR",
				Name: nil,
			},
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, formatGraphQLType(tc.input))
		})
	}
}

// TestConvertGraphQLTypeToJSONSchema ...
// Summary: TestConvertGraphQLTypeToJSONSchema
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Basic check for nil
	res := convertGraphQLTypeToJSONSchema(nil)
	require.NotNil(t, res)
	assert.Equal(t, "object", res.GetStructValue().Fields["type"].GetStringValue())

	// Basic check for recursion
	listType := &graphQLType{
		Kind: "LIST",
		OfType: &graphQLType{
			Kind: "SCALAR",
			Name: proto.String("String"),
		},
	}
	res = convertGraphQLTypeToJSONSchema(listType)
	assert.Equal(t, "array", res.GetStructValue().Fields["type"].GetStringValue())
	assert.Equal(t, "string", res.GetStructValue().Fields["items"].GetStructValue().Fields["type"].GetStringValue())
}

// TestGraphQLUpstream_Register_AddToolError ...
// Summary: TestGraphQLUpstream_Register_AddToolError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	u := NewGraphQLUpstream()
	// Mock the server to return valid schema
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
	}))
	defer server.Close()

	tm := tool.NewMockManagerInterface(ctrl)
	tm.EXPECT().AddTool(gomock.Any()).Return(errors.New("add tool error"))

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		GraphqlService: configv1.GraphQLUpstreamService_builder{
			Address: proto.String(server.URL),
		}.Build(),
	}.Build()

	_, _, _, err := u.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add tool")
}

// TestCallable_Call_NewRequestError ...
// Summary: TestCallable_Call_NewRequestError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	c := &Callable{
		address:       "http://example.com\x7f", // Invalid char
		authenticator: &dummyAuthenticator{},
		query:         "query { foo }",
	}

	req := &tool.ExecutionRequest{Arguments: map[string]interface{}{}}
	_, err := c.Call(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create dummy request")
}

type dummyAuthenticator struct{}

// Authenticate ...
// Summary: Authenticate
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}
