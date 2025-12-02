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
			input: "http\nmy-service\nhttp://localhost:8080\nget_user\nGet user by ID\nHTTP_METHOD_GET\n/users/{userId}\n",
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
			name:  "grpc",
			input: "grpc\nmy-grpc-service\nlocalhost:50051\ntrue\n",
			expected: `upstreamServices:
  - name: "my-grpc-service"
    grpcService:
      address: "localhost:50051"
      reflection:
        enabled: true`,
		},
		{
			name:  "openapi",
			input: "openapi\nmy-openapi-service\n./openapi.json\n",
			expected: `upstreamServices:
  - name: "my-openapi-service"
    openapiService:
      spec:
        path: "./openapi.json"`,
		},
		{
			name:  "graphql",
			input: "graphql\nmy-graphql-service\nhttp://localhost:8080/graphql\nuser\n{ id name }\n",
			expected: `upstreamServices:
  - name: "my-graphql-service"
    graphqlService:
      address: "http://localhost:8080/graphql"
      calls:
        - name: "user"
          selectionSet: "{ id name }"`,
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
