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
	"testing"

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
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"name": "test-tool", "description": "A test tool"}`)
	}))
	defer mockService.Close()

	// Register the mock service
	RegisterHTTPService(t, serverInfo.RegistrationClient, "test-service", mockService.URL, "test-tool", "/", "GET", nil)

	// List tools and verify the mock service's tool is present
	tools, err := serverInfo.ListTools(context.Background())
	require.NoError(t, err)
	require.Len(t, tools.Tools, 1)
	require.Equal(t, "test-tool", tools.Tools[0].Name)

	// Call the tool
	callParams := &mcp.CallToolParams{
		Name: "test-tool",
	}
	result, err := serverInfo.CallTool(context.Background(), callParams)
	require.NoError(t, err)
	require.NotNil(t, result)
}
