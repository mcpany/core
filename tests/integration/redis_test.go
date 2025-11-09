/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestRedisBus_ExternalServer(t *testing.T) {
	if !IsDockerSocketAccessible() {
		t.Skip("Docker is not available, skipping test")
	}
	redisAddr, redisCleanup := StartRedisContainer(t)
	defer redisCleanup()

	configContent := fmt.Sprintf(`
global_settings:
  message_bus:
    redis:
      address: "%s"
`, redisAddr)
	serverInfo := StartMCPANYServerWithConfig(t, "redis-test", configContent)
	defer serverInfo.CleanupFunc()

	// Mock service that registers a tool
	mockService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"name": "test-tool", "description": "A test tool"}`)
	}))
	defer mockService.Close()

	// Create a temporary OpenAPI spec file
	openapiSpec := fmt.Sprintf(`
openapi: 3.0.0
info:
  title: Test Service
  version: 1.0.0
servers:
  - url: %s
paths:
  /:
    get:
      operationId: test-tool
      responses:
        '200':
          description: OK
`, mockService.URL)

	specFile, err := os.CreateTemp(t.TempDir(), "openapi-*.yaml")
	require.NoError(t, err)
	_, err = specFile.WriteString(openapiSpec)
	require.NoError(t, err)
	err = specFile.Close()
	require.NoError(t, err)

	// Register the mock service using the OpenAPI spec
	RegisterOpenAPIService(t, serverInfo.RegistrationClient, "test-service", specFile.Name(), "", nil)

	// Poll until the tool is registered
	require.Eventually(t, func() bool {
		tools, err := serverInfo.ListTools(context.Background())
		if err != nil {
			// This is expected if the method isn't registered yet
			return false
		}
		for _, tool := range tools.Tools {
			if tool.Name == "test-tool" {
				return true
			}
		}
		return false
	}, 10*time.Second, 250*time.Millisecond, "Tool 'test-tool' was not registered in time")

	// Call the tool
	callParams := &mcp.CallToolParams{
		Name: "test-tool",
	}
	result, err := serverInfo.CallTool(context.Background(), callParams)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotEmpty(t, result.Content)
	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected TextContent")
	require.Contains(t, textContent.Text, `"name": "test-tool"`)
}
