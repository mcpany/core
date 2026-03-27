// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"context"
	"fmt"
	"testing"

	"github.com/mcpany/core/server/tests/integration"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestGRPCHelperFunctions(t *testing.T) {
	t.Parallel()
	t.Run("unauthenticated", func(t *testing.T) {
		t.Parallel()
		proc := BuildGRPCWeatherServer(t)
		require.NotNil(t, proc)
		err := proc.Start()
		require.NoError(t, err)
		defer proc.Stop()

		proc.Port = WaitForPort(t, proc)
		integration.WaitForTCPPort(t, proc.Port, integration.ServiceStartupTimeout)


		mcpanyTestServerInfo := integration.StartMCPANYServer(t, "grpc-weather-test")
		defer mcpanyTestServerInfo.CleanupFunc()

		upstreamEndpoint := fmt.Sprintf("127.0.0.1:%d", proc.Port)
		RegisterGRPCWeatherService(t, mcpanyTestServerInfo.RegistrationClient, upstreamEndpoint)

		// Verify the tool is registered
		ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
		defer cancel()

		client := mcp.NewClient(&mcp.Implementation{Name: "e2e-test-client"}, nil)
		transport := &mcp.StreamableClientTransport{
			Endpoint: mcpanyTestServerInfo.HTTPEndpoint,
		}
		session, err := client.Connect(ctx, transport, nil)
		require.NoError(t, err)
		defer func() { _ = session.Close() }()

		tools, err := session.ListTools(ctx, &mcp.ListToolsParams{})
		require.NoError(t, err)

		found := false
		for _, tool := range tools.Tools {
			// The tool name is derived from the service name and the RPC method name.
			if tool.Name == "e2e_grpc_weather.GetWeather" {
				found = true
				break
			}
		}
		require.True(t, found, "GetWeather tool should be registered for unauthenticated service")
	})

	t.Run("authenticated", func(t *testing.T) {
		t.Parallel()
		proc := BuildGRPCAuthedWeatherServer(t)
		require.NotNil(t, proc)
		err := proc.Start()
		require.NoError(t, err)
		defer proc.Stop()

		proc.Port = WaitForPort(t, proc)
		integration.WaitForTCPPort(t, proc.Port, integration.ServiceStartupTimeout)


		mcpanyTestServerInfo := integration.StartMCPANYServer(t, "grpc-authed-weather-test")
		defer mcpanyTestServerInfo.CleanupFunc()

		upstreamEndpoint := fmt.Sprintf("127.0.0.1:%d", proc.Port)
		RegisterGRPCAuthedWeatherService(t, mcpanyTestServerInfo.RegistrationClient, upstreamEndpoint)

		// Verify the tool is registered
		ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
		defer cancel()

		client := mcp.NewClient(&mcp.Implementation{Name: "e2e-test-client"}, nil)
		transport := &mcp.StreamableClientTransport{
			Endpoint: mcpanyTestServerInfo.HTTPEndpoint,
		}
		session, err := client.Connect(ctx, transport, nil)
		require.NoError(t, err)
		defer func() { _ = session.Close() }()

		tools, err := session.ListTools(ctx, &mcp.ListToolsParams{})
		require.NoError(t, err)

		found := false
		for _, tool := range tools.Tools {
			// The tool name is derived from the service name and the RPC method name.
			if tool.Name == "e2e_grpc_authed_weather.GetWeather" {
				found = true
				break
			}
		}
		require.True(t, found, "GetWeather tool should be registered for authenticated service")
	})
}
