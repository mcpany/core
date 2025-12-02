//go:build integration
// +build integration

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
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetrics(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpany-integration-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configContent := `
upstreamServices:
  - name: "http-echo-server"
    httpService:
      address: "http://localhost:8080"
      calls:
        - operationId: "echo"
          description: "Echoes back the request body"
          method: "HTTP_METHOD_POST"
          endpointPath: "/echo"
`
	configPath := filepath.Join(tempDir, "config.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mcpany := testutil.NewMCPAny(t, ctx,
		"--config-path", configPath,
		"--metrics-listen-address", "localhost:9090",
	)
	defer mcpany.Stop()

	// Wait for the server to be ready
	time.Sleep(2 * time.Second)

	// Make a request to the echo tool
	_, err = testutil.CallTool(t, "http-echo-server/-/echo", `{"message": "hello"}`)
	require.NoError(t, err)

	// Make a request to the metrics endpoint
	resp, err := http.Get("http://localhost:9090/metrics")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Check if the metrics are present in the response
	assert.Contains(t, string(body), "mcpany_tool_http_echo_server_echo_call_total 1")
	assert.Contains(t, string(body), "mcpany_tools_call_total 1")
}
