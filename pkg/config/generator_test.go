// Copyright 2025 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law_ or agreed to in writing, software
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
	"testing"
)

func TestGenerator_Generate(t *testing.T) {
	testCases := []struct {
		name      string
		inputs    []string
		expected  string
		expectErr bool
	}{
		{
			name: "HTTP Service",
			inputs: []string{
				"http",
				"test-http",
				"http://localhost:8080",
				"get_user",
				"Get user by ID",
				"HTTP_METHOD_GET",
				"/users/{userId}",
			},
			expected: `upstreamServices:
  - name: "test-http"
    httpService:
      address: "http://localhost:8080"
      calls:
        - operationId: "get_user"
          description: "Get user by ID"
          method: "HTTP_METHOD_GET"
          endpointPath: "/users/{userId}"
`,
		},
		{
			name: "gRPC Service",
			inputs: []string{
				"grpc",
				"test-grpc",
				"localhost:50051",
				"true",
			},
			expected: `upstreamServices:
  - name: "test-grpc"
    grpcService:
      address: "localhost:50051"
      reflection:
        enabled: true
`,
		},
		{
			name: "gRPC Service with reflection disabled",
			inputs: []string{
				"grpc",
				"test-grpc-disabled",
				"localhost:50051",
				"false",
			},
			expected: `upstreamServices:
  - name: "test-grpc-disabled"
    grpcService:
      address: "localhost:50051"
      reflection:
        enabled: false
`,
		},
		{
			name: "gRPC Service with invalid reflection input",
			inputs: []string{
				"grpc",
				"test-grpc-invalid",
				"localhost:50051",
				"invalid",
				"true",
			},
			expected: `upstreamServices:
  - name: "test-grpc-invalid"
    grpcService:
      address: "localhost:50051"
      reflection:
        enabled: true
`,
		},
		{
			name: "OpenAPI Service",
			inputs: []string{
				"openapi",
				"test-openapi",
				"./openapi.json",
			},
			expected: `upstreamServices:
  - name: "test-openapi"
    openapiService:
      spec:
        path: "./openapi.json"
`,
		},
		{
			name: "GraphQL Service",
			inputs: []string{
				"graphql",
				"test-graphql",
				"http://localhost:8080/graphql",
				"user",
				"{ id name }",
			},
			expected: `upstreamServices:
  - name: "test-graphql"
    graphqlService:
      address: "http://localhost:8080/graphql"
      calls:
        - name: "user"
          selectionSet: "{ id name }"
`,
		},
		{
			name:      "Unsupported Service Type",
			inputs:    []string{"invalid"},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var stdin bytes.Buffer
			for _, input := range tc.inputs {
				stdin.WriteString(input + "\n")
			}

			g := &Generator{
				Reader: bufio.NewReader(&stdin),
			}

			configData, err := g.Generate()

			if (err != nil) != tc.expectErr {
				t.Fatalf("Generate() error = %v, expectErr %v", err, tc.expectErr)
			}

			if !bytes.Equal(configData, []byte(tc.expected)) {
				t.Errorf("Generate() = %v, want %v", string(configData), tc.expected)
			}
		})
	}
}
