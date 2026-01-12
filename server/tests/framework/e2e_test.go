// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestGRPCHelperFunctions(t *testing.T) {
	// t.Parallel()
	t.Run("unauthenticated", func(t *testing.T) {
		// t.Parallel()
		var proc *integration.ManagedProcess
		var err error
		for i := 0; i < 5; i++ {
			proc = BuildGRPCWeatherServer(t)
			require.NotNil(t, proc)
			err = proc.Start()
			if err == nil {
				// Verify it's actually running and listening
				err = integration.WaitForTCPPortE(t, proc.Port, integration.ServiceStartupTimeout)
				if err == nil {
					break
				}
				// Stop the failed process before retrying
				proc.Stop()
				t.Logf("Server failed to listen (attempt %d/5): %v. Retrying...", i+1, err)
				err = fmt.Errorf("failed to listen: %w", err) // Ensure err is set for final check
			} else {
				t.Logf("Failed to start server (attempt %d/5): %v", i+1, err)
			}
			time.Sleep(100 * time.Millisecond)
		}
		require.NoError(t, err)
		defer proc.Stop()

		mcpanyTestServerInfo := integration.StartMCPANYServer(t, "grpc-weather-test")
		defer mcpanyTestServerInfo.CleanupFunc()

		upstreamEndpoint := fmt.Sprintf("localhost:%d", proc.Port)
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
		// t.Parallel()
		var proc *integration.ManagedProcess
		var err error
		for i := 0; i < 5; i++ {
			proc = BuildGRPCAuthedWeatherServer(t)
			require.NotNil(t, proc)
			err = proc.Start()
			if err == nil {
				// Verify it's actually running and listening
				err = integration.WaitForTCPPortE(t, proc.Port, integration.ServiceStartupTimeout)
				if err == nil {
					break
				}
				// Stop the failed process before retrying
				proc.Stop()
				t.Logf("Server failed to listen (attempt %d/5): %v. Retrying...", i+1, err)
				err = fmt.Errorf("failed to listen: %w", err) // Ensure err is set for final check
			} else {
				t.Logf("Failed to start server (attempt %d/5): %v", i+1, err)
			}
			time.Sleep(100 * time.Millisecond)
		}
		require.NoError(t, err)
		defer proc.Stop()

		mcpanyTestServerInfo := integration.StartMCPANYServer(t, "grpc-authed-weather-test")
		defer mcpanyTestServerInfo.CleanupFunc()

		upstreamEndpoint := fmt.Sprintf("localhost:%d", proc.Port)
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
