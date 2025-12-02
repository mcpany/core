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

package e2e

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/mcpany/core/pkg/config"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestCLI(t *testing.T) {
	t.Run("generate http service", func(t *testing.T) {
		cmd := exec.Command("go", "run", "../../cmd/mcp-any-cli/main.go")

		input := "http\ntest-http-service\nhttp://localhost:8080\napiKey\nX-API-Key\nmy-api-key\ny\nget-user\nGet a user\nGET\n/users/{id}\ny\nid\nuserId\nn\nn\n"
		cmd.Stdin = strings.NewReader(input)

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("unexpected error: %v\noutput: %s", err, string(output))
		}

		outputParts := strings.Split(string(output), "Generated configuration:")
		if len(outputParts) != 2 {
			t.Fatalf("unexpected output format: %s", string(output))
		}

		var cfg config.Config
		err = yaml.Unmarshal([]byte(outputParts[1]), &cfg)
		if err != nil {
			t.Fatalf("failed to unmarshal YAML: %v", err)
		}

		assert.Len(t, cfg.UpstreamServices, 1)
		service := cfg.UpstreamServices[0]
		assert.Equal(t, "test-http-service", service.Name)
		assert.NotNil(t, service.HTTPService)
		assert.Equal(t, "http://localhost:8080", service.HTTPService.Address)
		assert.Len(t, service.HTTPService.Calls, 1)
		call := service.HTTPService.Calls[0]
		assert.Equal(t, "get-user", call.OperationID)
		assert.Equal(t, "Get a user", call.Description)
		assert.Equal(t, "GET", call.Method)
		assert.Equal(t, "/users/{id}", call.EndpointPath)
		assert.Len(t, call.ParameterMappings, 1)
		param := call.ParameterMappings[0]
		assert.Equal(t, "id", param.InputParameterName)
		assert.Equal(t, "userId", param.TargetParameterName)
		assert.NotNil(t, service.UpstreamAuthentication)
		assert.NotNil(t, service.UpstreamAuthentication.APIKey)
		assert.Equal(t, "X-API-Key", service.UpstreamAuthentication.APIKey.HeaderName)
		assert.Equal(t, "my-api-key", service.UpstreamAuthentication.APIKey.APIKey.PlainText)
	})

	t.Run("generate grpc service", func(t *testing.T) {
		cmd := exec.Command("go", "run", "../../cmd/mcp-any-cli/main.go")

		input := "grpc\ntest-grpc-service\nlocalhost:50051\ny\n"
		cmd.Stdin = strings.NewReader(input)

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("unexpected error: %v\noutput: %s", err, string(output))
		}

		outputParts := strings.Split(string(output), "Generated configuration:")
		if len(outputParts) != 2 {
			t.Fatalf("unexpected output format: %s", string(output))
		}

		var cfg config.Config
		err = yaml.Unmarshal([]byte(outputParts[1]), &cfg)
		if err != nil {
			t.Fatalf("failed to unmarshal YAML: %v", err)
		}

		assert.Len(t, cfg.UpstreamServices, 1)
		service := cfg.UpstreamServices[0]
		assert.Equal(t, "test-grpc-service", service.Name)
		assert.NotNil(t, service.GRPCService)
		assert.Equal(t, "localhost:50051", service.GRPCService.Address)
		assert.True(t, service.GRPCService.Reflection.Enabled)
	})
}
