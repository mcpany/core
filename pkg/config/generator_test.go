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
	"bytes"
	"strings"
	"testing"
)

func TestGenerator(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		expected      string
		expectError   bool
		expectedError string
	}{
		{
			name: "HTTP Service without Auth",
			input: "http\n" +
				"my-http-service\n" +
				"http://localhost:8080\n" +
				"get_user\n" +
				"Get a user\n" +
				"HTTP_METHOD_GET\n" +
				"/users/123\n" +
				"no\n",
			expected: `upstreamServices:
  - name: "my-http-service"
    httpService:
      address: "http://localhost:8080"
      calls:
        - operationId: "get_user"
          description: "Get a user"
          method: "HTTP_METHOD_GET"
          endpointPath: "/users/123"
`,
		},
		{
			name: "HTTP Service with API Key Auth",
			input: "http\n" +
				"my-http-service\n" +
				"http://localhost:8080\n" +
				"get_user\n" +
				"Get a user\n" +
				"HTTP_METHOD_GET\n" +
				"/users/123\n" +
				"yes\n" +
				"apiKey\n" +
				"X-API-Key\n" +
				"my-secret-key\n",
			expected: `upstreamServices:
  - name: "my-http-service"
    httpService:
      address: "http://localhost:8080"
      calls:
        - operationId: "get_user"
          description: "Get a user"
          method: "HTTP_METHOD_GET"
          endpointPath: "/users/123"
    upstreamAuthentication:
      apiKey:
        headerName: "X-API-Key"
        apiKey:
          plainText: "my-secret-key"
`,
		},
		{
			name: "gRPC Service without Auth",
			input: "grpc\n" +
				"my-grpc-service\n" +
				"localhost:50051\n" +
				"true\n" +
				"no\n",
			expected: `upstreamServices:
  - name: "my-grpc-service"
    grpcService:
      address: "localhost:50051"
      reflection:
        enabled: true
`,
		},
		{
			name: "gRPC Service with Bearer Token Auth",
			input: "grpc\n" +
				"my-grpc-service\n" +
				"localhost:50051\n" +
				"true\n" +
				"yes\n" +
				"bearerToken\n" +
				"my-bearer-token\n",
			expected: `upstreamServices:
  - name: "my-grpc-service"
    grpcService:
      address: "localhost:50051"
      reflection:
        enabled: true
    upstreamAuthentication:
      bearerToken:
        token:
          plainText: "my-bearer-token"
`,
		},
		{
			name: "OpenAPI Service without Auth",
			input: "openapi\n" +
				"my-openapi-service\n" +
				"./openapi.json\n" +
				"no\n",
			expected: `upstreamServices:
  - name: "my-openapi-service"
    openapiService:
      spec:
        path: "./openapi.json"
`,
		},
		{
			name: "OpenAPI Service with Basic Auth",
			input: "openapi\n" +
				"my-openapi-service\n" +
				"./openapi.json\n" +
				"yes\n" +
				"basicAuth\n" +
				"my-user\n" +
				"my-password\n",
			expected: `upstreamServices:
  - name: "my-openapi-service"
    openapiService:
      spec:
        path: "./openapi.json"
    upstreamAuthentication:
      basicAuth:
        username: "my-user"
        password:
          plainText: "my-password"
`,
		},
		{
			name: "GraphQL Service without Auth",
			input: "graphql\n" +
				"my-graphql-service\n" +
				"http://localhost:8081/graphql\n" +
				"user\n" +
				"{ id name }\n" +
				"no\n",
			expected: `upstreamServices:
  - name: "my-graphql-service"
    graphqlService:
      address: "http://localhost:8081/graphql"
      calls:
        - name: "user"
          selectionSet: "{ id name }"
`,
		},
		{
			name: "GraphQL Service with Basic Auth",
			input: "graphql\n" +
				"my-graphql-service\n" +
				"http://localhost:8081/graphql\n" +
				"user\n" +
				"{ id name }\n" +
				"yes\n" +
				"basicAuth\n" +
				"my-user\n" +
				"my-password\n",
			expected: `upstreamServices:
  - name: "my-graphql-service"
    graphqlService:
      address: "http://localhost:8081/graphql"
      calls:
        - name: "user"
          selectionSet: "{ id name }"
    upstreamAuthentication:
      basicAuth:
        username: "my-user"
        password:
          plainText: "my-password"
`,
		},
		{
			name:          "Invalid Service Type",
			input:         "invalid\n",
			expectError:   true,
			expectedError: "unsupported service type: invalid",
		},
		{
			name: "Empty Service Name with recovery",
			input: "http\n\nmy-service\n" +
				"http://localhost:8080\n" +
				"get_user\n" +
				"Get a user\n" +
				"HTTP_METHOD_GET\n" +
				"/users/123\n" +
				"no\n",
			expected: `upstreamServices:
  - name: "my-service"
    httpService:
      address: "http://localhost:8080"
      calls:
        - operationId: "get_user"
          description: "Get a user"
          method: "HTTP_METHOD_GET"
          endpointPath: "/users/123"
`,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tc.input))
			generator := &Generator{Reader: reader}

			output, err := generator.Generate()

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				} else if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("Expected error to contain '%s', but got '%s'", tc.expectedError, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !bytes.Equal(bytes.TrimSpace(output), bytes.TrimSpace([]byte(tc.expected))) {
				t.Errorf("Expected:\n%s\n\nGot:\n%s", tc.expected, output)
			}
		})
	}
}
