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
	"io"
	"os"
	"strings"
	"testing"
)

func TestGenerator(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		expected      string
		expectedError string
	}{
		{
			name: "HTTP Service",
			input: "http\nmy-http-service\nhttps://api.example.com\nget_user\nGet user by ID\nHTTP_METHOD_GET\n/users/{userId}\n",
			expected: `upstreamServices:
  - name: "my-http-service"
    httpService:
      address: "https://api.example.com"
      calls:
        - operationId: "get_user"
          description: "Get user by ID"
          method: "HTTP_METHOD_GET"
          endpointPath: "/users/{userId}"
`,
		},
		{
			name: "gRPC Service",
			input: "grpc\nmy-grpc-service\nlocalhost:50052\ntrue\n",
			expected: `upstreamServices:
  - name: "my-grpc-service"
    grpcService:
      address: "localhost:50052"
      reflection:
        enabled: true
`,
		},
		{
			name: "OpenAPI Service",
			input: "openapi\nmy-openapi-service\n./openapi.json\n",
			expected: `upstreamServices:
  - name: "my-openapi-service"
    openapiService:
      spec:
        path: "./openapi.json"
`,
		},
		{
			name: "GraphQL Service",
			input: "graphql\nmy-graphql-service\nhttp://localhost:8080/graphql\nuser\n{ id name }\n",
			expected: `upstreamServices:
  - name: "my-graphql-service"
    graphqlService:
      address: "http://localhost:8080/graphql"
      calls:
        - name: "user"
          selectionSet: "{ id name }"
`,
		},
		{
			name:          "Unsupported Service Type",
			input:         "invalid\n",
			expectedError: "unsupported service type: invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := strings.NewReader(tc.input)
			oldStdin := os.Stdin
			defer func() { os.Stdin = oldStdin }()
			r, w, _ := os.Pipe()
			os.Stdin = r
			io.Copy(w, input)
			w.Close()

			generator := NewGenerator()
			output, err := generator.Generate()

			if tc.expectedError != "" {
				if err == nil || !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("expected error containing %q, got %v", tc.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if strings.TrimSpace(string(output)) != strings.TrimSpace(tc.expected) {
				t.Errorf("expected:\n%s\ngot:\n%s", tc.expected, string(output))
			}
		})
	}
}
