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
				"http",                    // Service Type
				"test-http",               // Name
				"http://localhost:8080",   // Address
				"get_user",                // OpID
				"Get user by ID",          // Desc
				"HTTP_METHOD_GET",         // Method
				"/users/{userId}",         // Path
				"n",                       // Add another? No
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
				"grpc",            // Type
				"test-grpc",       // Name
				"localhost:50051", // Address
				"true",            // Reflection
				"n",               // Add another? No
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
			name: "Command Service",
			inputs: []string{
				"command",       // Type
				"test-cmd",      // Name
				"python script.py", // Command
				"arg1 arg2",     // Args
				"y",             // Add Env? Yes
				"API_KEY",       // Key
				"12345",         // Value
				"n",             // Add Env? No
				"n",             // Add another service? No
			},
			expected: `upstreamServices:
  - name: "test-cmd"
    command:
      command: "python script.py"
      args:
        - "arg1"
        - "arg2"
      env:
        API_KEY: "12345"
`,
		},
		{
			name: "Multiple Services",
			inputs: []string{
				"http",
				"service-1",
				"http://s1",
				"op1",
				"desc1",
				"GET",
				"/path1",
				"y", // Add another? Yes
				"grpc",
				"service-2",
				"localhost:9090",
				"false",
				"n", // Add another? No
			},
			expected: `upstreamServices:
  - name: "service-1"
    httpService:
      address: "http://s1"
      calls:
        - operationId: "op1"
          description: "desc1"
          method: "HTTP_METHOD_GET"
          endpointPath: "/path1"
  - name: "service-2"
    grpcService:
      address: "localhost:9090"
      reflection:
        enabled: false
`,
		},
		{
			name:      "Unsupported Service Type",
			inputs:    []string{"invalid", "done"}, // Invalid retry, then done
			expected:  "upstreamServices:\n",       // Empty config
			expectErr: false,                       // Should not error, just loop
		},
		{
			name: "Done immediately",
			inputs: []string{
				"done",
			},
			expected: "upstreamServices:\n",
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

			// Normalize line endings and trim space for comparison
			got := string(bytes.TrimSpace(configData))
			want := string(bytes.TrimSpace([]byte(tc.expected)))

			if got != want {
				t.Errorf("Generate() mismatch:\nGOT:\n%s\nWANT:\n%s", got, want)
			}
		})
	}
}

func TestNewGenerator(t *testing.T) {
	g := NewGenerator(nil)
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

	t.Run("Prompt Error during HTTP Service generation", func(t *testing.T) {
		// Test EOF at each step of HTTP service generation
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
}
