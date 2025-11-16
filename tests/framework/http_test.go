/*
 * Copyright 2024 Author(s) of MCP Any
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

package framework

import (
	"context"
	"fmt"
	"testing"

	"github.com/mcpany/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestHTTPHelperFunctions(t *testing.T) {
	t.Run("unauthenticated", func(t *testing.T) {
		t.Parallel()
		proc := BuildHTTPEchoServer(t)
		require.NotNil(t, proc)
		err := proc.Start()
		require.NoError(t, err)
		defer proc.Stop()

		integration.WaitForTCPPort(t, proc.Port, integration.ServiceStartupTimeout)

		mcpanyTestServerInfo := integration.StartMCPANYServer(t, "http-echo-test")
		defer mcpanyTestServerInfo.CleanupFunc()

		upstreamEndpoint := fmt.Sprintf("http://localhost:%d", proc.Port)
		RegisterHTTPEchoService(t, mcpanyTestServerInfo.RegistrationClient, upstreamEndpoint)

		// Verify the tool is registered
		ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
		defer cancel()

		client := mcp.NewClient(&mcp.Implementation{Name: "e2e-test-client"}, nil)
		transport := &mcp.StreamableClientTransport{
			Endpoint: mcpanyTestServerInfo.HTTPEndpoint,
		}
		session, err := client.Connect(ctx, transport, nil)
		require.NoError(t, err)
		defer session.Close()

		tools, err := session.ListTools(ctx, &mcp.ListToolsParams{})
		require.NoError(t, err)

		found := false
		for _, tool := range tools.Tools {
			if tool.Name == "e2e_http_echo.echo" {
				found = true
				break
			}
		}
		require.True(t, found, "echo tool should be registered for unauthenticated service")
	})

	t.Run("authenticated", func(t *testing.T) {
		t.Parallel()
		proc := BuildHTTPAuthedEchoServer(t)
		require.NotNil(t, proc)
		err := proc.Start()
		require.NoError(t, err)
		defer proc.Stop()

		integration.WaitForTCPPort(t, proc.Port, integration.ServiceStartupTimeout)

		mcpanyTestServerInfo := integration.StartMCPANYServer(t, "http-authed-echo-test")
		defer mcpanyTestServerInfo.CleanupFunc()

		upstreamEndpoint := fmt.Sprintf("http://localhost:%d", proc.Port)
		RegisterHTTPAuthedEchoService(t, mcpanyTestServerInfo.RegistrationClient, upstreamEndpoint)

		// Verify the tool is registered
		ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
		defer cancel()

		client := mcp.NewClient(&mcp.Implementation{Name: "e2e-test-client"}, nil)
		transport := &mcp.StreamableClientTransport{
			Endpoint: mcpanyTestServerInfo.HTTPEndpoint,
		}
		session, err := client.Connect(ctx, transport, nil)
		require.NoError(t, err)
		defer session.Close()

		tools, err := session.ListTools(ctx, &mcp.ListToolsParams{})
		require.NoError(t, err)

		found := false
		for _, tool := range tools.Tools {
			if tool.Name == "e2e_http_authed_echo.echo" {
				found = true
				break
			}
		}
		require.True(t, found, "echo tool should be registered for authenticated service")
	})
}
