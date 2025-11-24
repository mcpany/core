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

package framework

import (
	"context"
	"fmt"
	"testing"

	"github.com/mcpany/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestEverythingHelperFunctions(t *testing.T) {
	t.Parallel()
	proc := BuildEverythingServer(t)
	require.NotNil(t, proc)
	err := proc.Start()
	require.NoError(t, err)
	defer proc.Stop()

	integration.WaitForTCPPort(t, proc.Port, integration.ServiceStartupTimeout)

	mcpanyTestServerInfo := integration.StartMCPANYServer(t, "everything-test")
	defer mcpanyTestServerInfo.CleanupFunc()

	upstreamEndpoint := fmt.Sprintf("http://localhost:%d/mcp", proc.Port)
	RegisterEverythingService(t, mcpanyTestServerInfo.RegistrationClient, upstreamEndpoint)

	// Verify that at least one tool is registered
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

	require.Greater(t, len(tools.Tools), 0, "at least one tool should be registered from the everything server")
}
