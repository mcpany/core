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

package e2e_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/mcpserver"
	"github.com/mcpany/core/pkg/testutil"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestE2E_HealthCheck(t *testing.T) {
	// Create a mock upstream service that can be healthy or unhealthy.
	healthy := true
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			if healthy {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
			}
			return
		}
		fmt.Fprintln(w, "ok")
	}))
	defer upstream.Close()

	// Create a config file for the server.
	configFile := `
upstreamServices:
  - name: "my-http-service"
    httpService:
      address: "` + upstream.URL + `"
      healthCheck:
        endpoint: "/healthz"
        expectedCode: 200
        interval: "1s"
        timeout: "1s"
      calls:
        - operationId: "get_user"
          description: "Get user by ID"
          method: "HTTP_METHOD_GET"
          endpointPath: "/users/{userId}"
`
	configDir, err := os.MkdirTemp("", "e2e-health-check")
	require.NoError(t, err)
	defer os.RemoveAll(configDir)

	configPath := filepath.Join(configDir, "config.yaml")
	err = os.WriteFile(configPath, []byte(configFile), 0644)
	require.NoError(t, err)

	// Start the MCP Any server.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, mcpServer, err := testutil.StartTestServer(ctx, t, []string{configPath})
	require.NoError(t, err)

	client, clientSession := testutil.CreateTestClient(t, ctx, mcpServer)
	defer clientSession.Close()

	// Wait for the first health check to run.
	time.Sleep(2 * time.Second)

	// The service should be healthy.
	_, err = client.CallTool(ctx, &mcp.CallToolParams{
		Name: "my-http-service/-/get_user",
		Arguments: map[string]interface{}{
			"userId": "123",
		},
	})
	require.NoError(t, err)

	// Make the upstream service unhealthy.
	healthy = false
	time.Sleep(2 * time.Second)

	// The service should be unhealthy.
	_, err = client.CallTool(ctx, &mcp.CallToolParams{
		Name: "my-http-service/-/get_user",
		Arguments: map[string]interface{}{
			"userId": "123",
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), `service "my-http-service" is unhealthy`)

	// Make the upstream service healthy again.
	healthy = true
	time.Sleep(2 * time.Second)

	// The service should be healthy again.
	_, err = client.CallTool(ctx, &mcp.CallToolParams{
		Name: "my-http-service/-/get_user",
		Arguments: map[string]interface{}{
			"userId": "123",
		},
	})
	require.NoError(t, err)
}
