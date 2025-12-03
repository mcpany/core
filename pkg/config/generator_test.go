// Copyright 2025 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"bufio"
	"strings"
	"testing"
)

func TestGenerator(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:  "http",
			input: "http\nmy-service\nhttp://localhost:8080\nget_user\nGet user by ID\nHTTP_METHOD_GET\n/users/{userId}\nno\n",
			expected: `upstreamServices:
  - name: "my-service"
    httpService:
      address: "http://localhost:8080"
      calls:
        - operationId: "get_user"
          description: "Get user by ID"
          method: "HTTP_METHOD_GET"
          endpointPath: "/users/{userId}"`,
		},
		{
			name:  "http_with_auth_plain_text",
			input: "http\nmy-service\nhttp://localhost:8080\nget_user\nGet user by ID\nHTTP_METHOD_GET\n/users/{userId}\nyes\napiKey\nX-API-Key\nplainText\nmy-secret-key\n",
			expected: `upstreamServices:
  - name: "my-service"
    httpService:
      address: "http://localhost:8080"
      calls:
        - operationId: "get_user"
          description: "Get user by ID"
          method: "HTTP_METHOD_GET"
          endpointPath: "/users/{userId}"
    upstreamAuthentication:
      apiKey:
        headerName: "X-API-Key"
        apiKey:
          plainText: "my-secret-key"`,
		},
		{
			name:  "http_with_auth_env_var",
			input: "http\nmy-service\nhttp://localhost:8080\nget_user\nGet user by ID\nHTTP_METHOD_GET\n/users/{userId}\nyes\napiKey\nX-API-Key\nenvironmentVariable\nMY_API_KEY\n",
			expected: `upstreamServices:
  - name: "my-service"
    httpService:
      address: "http://localhost:8080"
      calls:
        - operationId: "get_user"
          description: "Get user by ID"
          method: "HTTP_METHOD_GET"
          endpointPath: "/users/{userId}"
    upstreamAuthentication:
      apiKey:
        headerName: "X-API-Key"
        apiKey:
          environmentVariable: "MY_API_KEY"`,
		},
		{
			name:  "grpc",
			input: "grpc\nmy-grpc-service\nlocalhost:50051\ntrue\nno\n",
			expected: `upstreamServices:
  - name: "my-grpc-service"
    grpcService:
      address: "localhost:50051"
      reflection:
        enabled: true`,
		},
		{
			name:  "grpc_with_auth_plain_text",
			input: "grpc\nmy-grpc-service\nlocalhost:50051\ntrue\nyes\napiKey\nX-API-Key\nplainText\nmy-secret-key\n",
			expected: `upstreamServices:
  - name: "my-grpc-service"
    grpcService:
      address: "localhost:50051"
      reflection:
        enabled: true
    upstreamAuthentication:
      apiKey:
        headerName: "X-API-Key"
        apiKey:
          plainText: "my-secret-key"`,
		},
		{
			name:  "openapi",
			input: "openapi\nmy-openapi-service\n./openapi.json\nno\n",
			expected: `upstreamServices:
  - name: "my-openapi-service"
    openapiService:
      spec:
        path: "./openapi.json"`,
		},
		{
			name:  "openapi_with_auth_env_var",
			input: "openapi\nmy-openapi-service\n./openapi.json\nyes\napiKey\nX-API-Key\nenvironmentVariable\nMY_API_KEY\n",
			expected: `upstreamServices:
  - name: "my-openapi-service"
    openapiService:
      spec:
        path: "./openapi.json"
    upstreamAuthentication:
      apiKey:
        headerName: "X-API-Key"
        apiKey:
          environmentVariable: "MY_API_KEY"`,
		},
		{
			name:  "graphql",
			input: "graphql\nmy-graphql-service\nhttp://localhost:8080/graphql\nuser\n{ id name }\nno\n",
			expected: `upstreamServices:
  - name: "my-graphql-service"
    graphqlService:
      address: "http://localhost:8080/graphql"
      calls:
        - name: "user"
          selectionSet: "{ id name }"`,
		},
		{
			name:  "graphql_with_auth_plain_text",
			input: "graphql\nmy-graphql-service\nhttp://localhost:8080/graphql\nuser\n{ id name }\nyes\napiKey\nX-API-Key\nplainText\nmy-secret-key\n",
			expected: `upstreamServices:
  - name: "my-graphql-service"
    graphqlService:
      address: "http://localhost:8080/graphql"
      calls:
        - name: "user"
          selectionSet: "{ id name }"
    upstreamAuthentication:
      apiKey:
        headerName: "X-API-Key"
        apiKey:
          plainText: "my-secret-key"`,
		},
		{
			name:        "unsupported",
			input:       "foo\n",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tc.input))
			generator := &Generator{
				Reader: reader,
			}

			configData, err := generator.Generate()
			if tc.expectError {
				if err == nil {
					t.Fatal("expected an error, but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if strings.TrimSpace(string(configData)) != strings.TrimSpace(tc.expected) {
				t.Errorf("unexpected config data:\ngot:\n%s\nwant:\n%s", string(configData), tc.expected)
			}
		})
	}
}
