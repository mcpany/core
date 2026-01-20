// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"bufio"
	"bytes"
	"fmt"
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
				"http://127.0.0.1:8080",
				"get_user",
				"Get user by ID",
				"HTTP_METHOD_GET",
				"/users/{userId}",
			},
			expected: `upstreamServices:
  - name: "test-http"
    httpService:
      address: "http://127.0.0.1:8080"
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
				"127.0.0.1:50051",
				"true",
			},
			expected: `upstreamServices:
  - name: "test-grpc"
    grpcService:
      address: "127.0.0.1:50051"
      reflection:
        enabled: true
`,
		},
		{
			name: "gRPC Service with 'y' input",
			inputs: []string{
				"grpc",
				"test-grpc-y",
				"127.0.0.1:50051",
				"y",
			},
			expected: `upstreamServices:
  - name: "test-grpc-y"
    grpcService:
      address: "127.0.0.1:50051"
      reflection:
        enabled: true
`,
		},
		{
			name: "gRPC Service with 'N' input",
			inputs: []string{
				"grpc",
				"test-grpc-n",
				"127.0.0.1:50051",
				"N",
			},
			expected: `upstreamServices:
  - name: "test-grpc-n"
    grpcService:
      address: "127.0.0.1:50051"
      reflection:
        enabled: false
`,
		},
		{
			name: "HTTP Service Uppercase Type",
			inputs: []string{
				"HTTP",
				"test-http-upper",
				"http://127.0.0.1:8080",
				"get_user",
				"Get user by ID",
				"HTTP_METHOD_GET",
				"/users/{userId}",
			},
			expected: `upstreamServices:
  - name: "test-http-upper"
    httpService:
      address: "http://127.0.0.1:8080"
      calls:
        - operationId: "get_user"
          description: "Get user by ID"
          method: "HTTP_METHOD_GET"
          endpointPath: "/users/{userId}"
`,
		},
		{
			name: "gRPC Service with reflection disabled",
			inputs: []string{
				"grpc",
				"test-grpc-disabled",
				"127.0.0.1:50051",
				"false",
			},
			expected: `upstreamServices:
  - name: "test-grpc-disabled"
    grpcService:
      address: "127.0.0.1:50051"
      reflection:
        enabled: false
`,
		},
		{
			name: "gRPC Service with invalid reflection input",
			inputs: []string{
				"grpc",
				"test-grpc-invalid",
				"127.0.0.1:50051",
				"invalid",
				"true",
			},
			expected: `upstreamServices:
  - name: "test-grpc-invalid"
    grpcService:
      address: "127.0.0.1:50051"
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
				"http://127.0.0.1:8080/graphql",
				"user",
				"{ id name }",
			},
			expected: `upstreamServices:
  - name: "test-graphql"
    graphqlService:
      address: "http://127.0.0.1:8080/graphql"
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

func TestNewGenerator(t *testing.T) {
	g := NewGenerator()
	if g == nil {
		t.Fatal("NewGenerator() returned nil")
	}
	if g.Reader == nil {
		t.Error("NewGenerator().Reader is nil")
	}
}

type errorReader struct{}

func (r *errorReader) Read(_ []byte) (n int, err error) {
	return 0, fmt.Errorf("read error")
}

func (r *errorReader) ReadString(_ byte) (string, error) {
	return "", fmt.Errorf("read error")
}

func TestGenerator_Generate_Errors(t *testing.T) {
	t.Run("Prompt Error on Service Type", func(t *testing.T) {
		g := &Generator{
			Reader: bufio.NewReader(&errorReader{}),
		}
		_, err := g.Generate()
		if err == nil {
			t.Error("Expected error when prompt fails immediately")
		}
	})

	t.Run("EOF Error", func(t *testing.T) {
		// Empty input causes EOF immediately
		g := &Generator{
			Reader: bufio.NewReader(bytes.NewBufferString("")),
		}
		_, err := g.Generate()
		if err == nil {
			t.Error("Expected error on empty input")
		}
	})

	t.Run("Prompt Error during HTTP Service generation", func(t *testing.T) {
		// Test EOF at each step of HTTP service generation
		// Steps: Name, Address, OpID, Desc, Method, EndpointPath
		inputs := []string{
			"http\n",
			"http\nname\n",
			"http\nname\naddress\n",
			"http\nname\naddress\nopID\n",
			"http\nname\naddress\nopID\ndesc\n",
			"http\nname\naddress\nopID\ndesc\nGET\n",
		}

		for _, input := range inputs {
			g := &Generator{
				Reader: bufio.NewReader(bytes.NewBufferString(input)),
			}
			_, err := g.Generate()
			if err == nil {
				t.Errorf("Expected error for incomplete input: %q", input)
			}
		}
	})

	t.Run("Prompt Error during gRPC Service generation", func(t *testing.T) {
		g := &Generator{
			Reader: bufio.NewReader(bytes.NewBufferString("grpc\n")),
		}
		_, err := g.Generate()
		if err == nil {
			t.Error("Expected error when grpc service input is incomplete")
		}
	})

	t.Run("Prompt Error during gRPC Reflection loop", func(t *testing.T) {
		// grpc, name, address, then EOF
		input := "grpc\nname\naddress\n"
		g := &Generator{
			Reader: bufio.NewReader(bytes.NewBufferString(input)),
		}
		_, err := g.Generate()
		if err == nil {
			t.Error("Expected error when grpc reflection input is missing")
		}
	})

	t.Run("Prompt Error during OpenAPI Service generation", func(t *testing.T) {
		inputs := []string{
			"openapi\n",
			"openapi\nname\n",
		}
		for _, input := range inputs {
			g := &Generator{
				Reader: bufio.NewReader(bytes.NewBufferString(input)),
			}
			_, err := g.Generate()
			if err == nil {
				t.Errorf("Expected error for incomplete openapi input: %q", input)
			}
		}
	})

	t.Run("Prompt Error during GraphQL Service generation", func(t *testing.T) {
		inputs := []string{
			"graphql\n",
			"graphql\nname\n",
			"graphql\nname\naddress\n",
			"graphql\nname\naddress\ncallname\n",
		}
		for _, input := range inputs {
			g := &Generator{
				Reader: bufio.NewReader(bytes.NewBufferString(input)),
			}
			_, err := g.Generate()
			if err == nil {
				t.Errorf("Expected error for incomplete graphql input: %q", input)
			}
		}
	})
}

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

func TestGenerator_Prompt_Bug(t *testing.T) {
	// Input without a trailing newline
	input := "some input"
	reader := bufio.NewReader(bytes.NewBufferString(input))
	g := &Generator{
		Reader: reader,
	}

	result, err := g.prompt("prompt: ")

	// We expect "some input" but the current implementation returns "" and an error (probably EOF)
	// We want to handle EOF gracefully if there is content.

	if result != "some input" {
		t.Errorf("Expected 'some input', got '%s'", result)
	}

	// The current implementation returns an error on EOF even if there is data.
	// Depending on how we want to define "bug", usually tools should accept the last line even without newline.
	if err != nil {
		// It returns error because of EOF, but we might want to check if it returned the content at least?
		// Current code: return "", err
		t.Logf("Got error as expected from current implementation: %v", err)
	}
}
