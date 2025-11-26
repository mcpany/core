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
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/app"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestWatchMode(t *testing.T) {
	dir, err := os.MkdirTemp("", "watch_mode_test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	configFile := filepath.Join(dir, "config.yaml")
	err = os.WriteFile(configFile, []byte(`
upstreamServices:
  - name: "my-http-service"
    httpService:
      address: "https://api.example.com"
      calls:
        - operationId: "get_user"
          description: "Get user by ID"
          method: "HTTP_METHOD_GET"
          endpointPath: "/users/{userId}"
`), 0644)
	require.NoError(t, err)

	// Find a free port for the server to listen on.
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	addr := lis.Addr().String()
	lis.Close() // Close the listener so the app can use the port.

	fs := afero.NewOsFs()
	runner := app.NewApplication()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := runner.Run(
			ctx,
			fs,
			false,
			true,
			addr,
			"",
			[]string{configFile},
			5*time.Second,
		)
		require.NoError(t, err)
	}()

	// Wait for the server to start by polling the health check endpoint.
	require.Eventually(t, func() bool {
		err := app.HealthCheckWithContext(ctx, os.Stdout, addr)
		return err == nil
	}, 5*time.Second, 100*time.Millisecond, "server should start up")

	// Check that the tool is registered
	tools := runner.ListTools()
	require.Len(t, tools, 1)
	require.Equal(t, "my-http-service/-/get_user", tools[0].Tool().Name)

	// Modify the config file
	err = os.WriteFile(configFile, []byte(`
upstreamServices:
  - name: "my-http-service"
    httpService:
      address: "https://api.example.com"
      calls:
        - operationId: "get_user"
          description: "Get user by ID"
          method: "HTTP_METHOD_GET"
          endpointPath: "/users/{userId}"
        - operationId: "create_user"
          description: "Create a new user"
          method: "HTTP_METHOD_POST"
          endpointPath: "/users"
`), 0644)
	require.NoError(t, err)

	// Wait for the server to reload and register the new tool.
	require.Eventually(t, func() bool {
		tools := runner.ListTools()
		if len(tools) != 2 {
			return false
		}
		// The order of tools is not guaranteed, so we check for the presence of both.
		toolNames := make(map[string]bool)
		for _, tool := range tools {
			toolNames[*tool.Tool().Name] = true
		}
		return toolNames["my-http-service/-/get_user"] && toolNames["my-http-service/-/create_user"]
	}, 5*time.Second, 100*time.Millisecond, "server should reload and register the new tool")

	// Remove a service from the config file
	err = os.WriteFile(configFile, []byte(`
upstreamServices:
  - name: "my-http-service"
    httpService:
      address: "https://api.example.com"
      calls:
        - operationId: "create_user"
          description: "Create a new user"
          method: "HTTP_METHOD_POST"
          endpointPath: "/users"
`), 0644)
	require.NoError(t, err)

	// Wait for the server to reload and unregister the tool.
	require.Eventually(t, func() bool {
		tools := runner.ListTools()
		if len(tools) != 1 {
			return false
		}
		return *tools[0].Tool().Name == "my-http-service/-/create_user"
	}, 5*time.Second, 100*time.Millisecond, "server should reload and unregister the tool")
}
