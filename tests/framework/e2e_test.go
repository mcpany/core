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
	"time"

	"github.com/mcpany/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestGRPCHelperFunctions(t *testing.T) {
	t.Run("unauthenticated", func(t *testing.T) {
		t.Parallel()
		proc := BuildGRPCWeatherServer(t)
		require.NotNil(t, proc)
		err := proc.Start()
		require.NoError(t, err)
		defer proc.Stop()

		integration.WaitForTCPPort(t, proc.Port, integration.ServiceStartupTimeout)

		mcpanyTestServerInfo := integration.StartMCPANYServer(t, "grpc-weather-test")
		defer mcpanyTestServerInfo.CleanupFunc()

		upstreamEndpoint := fmt.Sprintf("localhost:%d", proc.Port)
		RegisterGRPCWeatherService(t, mcpanyTestServerInfo.RegistrationClient, upstreamEndpoint)

		verifyToolRegistered(t, mcpanyTestServerInfo.HTTPEndpoint, "e2e_grpc_weather.GetWeather", "GetWeather tool should be registered for unauthenticated service")
	})

	t.Run("authenticated", func(t *testing.T) {
		t.Parallel()
		proc := BuildGRPCAuthedWeatherServer(t)
		require.NotNil(t, proc)
		err := proc.Start()
		require.NoError(t, err)
		defer proc.Stop()

		integration.WaitForTCPPort(t, proc.Port, integration.ServiceStartupTimeout)

		mcpanyTestServerInfo := integration.StartMCPANYServer(t, "grpc-authed-weather-test")
		defer mcpanyTestServerInfo.CleanupFunc()

		upstreamEndpoint := fmt.Sprintf("localhost:%d", proc.Port)
		RegisterGRPCAuthedWeatherService(t, mcpanyTestServerInfo.RegistrationClient, upstreamEndpoint)

		verifyToolRegistered(t, mcpanyTestServerInfo.HTTPEndpoint, "e2e_grpc_authed_weather.GetWeather", "GetWeather tool should be registered for authenticated service")
	})
}

func TestWebsocketHelperFunctions(t *testing.T) {
	t.Parallel()
	proc := BuildWebsocketWeatherServer(t)
	require.NotNil(t, proc)
	err := proc.Start()
	require.NoError(t, err)
	defer proc.Stop()

	integration.WaitForTCPPort(t, 8091, integration.ServiceStartupTimeout)

	mcpanyTestServerInfo := integration.StartMCPANYServer(t, "websocket-weather-test")
	defer mcpanyTestServerInfo.CleanupFunc()

	upstreamEndpoint := fmt.Sprintf("ws://localhost:%d/echo", 8091)
	RegisterWebsocketWeatherService(t, mcpanyTestServerInfo.RegistrationClient, upstreamEndpoint)

	verifyToolRegistered(t, mcpanyTestServerInfo.HTTPEndpoint, "e2e_websocket_weather.GetWeather", "GetWeather tool should be registered for websocket service")
}

func TestWebrtcHelperFunctions(t *testing.T) {
	t.Parallel()
	proc := BuildWebrtcWeatherServer(t)
	require.NotNil(t, proc)
	err := proc.Start()
	require.NoError(t, err)
	defer proc.Stop()

	integration.WaitForTCPPort(t, proc.Port, integration.ServiceStartupTimeout)

	mcpanyTestServerInfo := integration.StartMCPANYServer(t, "webrtc-weather-test")
	defer mcpanyTestServerInfo.CleanupFunc()

	upstreamEndpoint := fmt.Sprintf("http://localhost:%d/signal", proc.Port)
	RegisterWebrtcWeatherService(t, mcpanyTestServerInfo.RegistrationClient, upstreamEndpoint)

	verifyToolRegistered(t, mcpanyTestServerInfo.HTTPEndpoint, "e2e_webrtc_weather.GetWeather", "GetWeather tool should be registered for webrtc service")
}

func TestStdioHelperFunctions(t *testing.T) {
	t.Parallel()
	BuildStdioServer(t)

	mcpanyTestServerInfo := integration.StartMCPANYServer(t, "stdio-test")
	defer mcpanyTestServerInfo.CleanupFunc()

	RegisterStdioService(t, mcpanyTestServerInfo.RegistrationClient, "")

	verifyToolRegistered(t, mcpanyTestServerInfo.HTTPEndpoint, "e2e_everything_server_stdio.time.getTime", "getTime tool should be registered for stdio service")
}

func TestStdioDockerHelperFunctions(t *testing.T) {
	t.Parallel()
	BuildStdioDockerServer(t)

	mcpanyTestServerInfo := integration.StartMCPANYServer(t, "stdio-docker-test")
	defer mcpanyTestServerInfo.CleanupFunc()

	RegisterStdioDockerService(t, mcpanyTestServerInfo.RegistrationClient, "")

	verifyToolRegistered(t, mcpanyTestServerInfo.HTTPEndpoint, "e2e-cowsay-server.cowsay", "cowsay tool should be registered for stdio docker service")
}

func TestOpenAPIHelperFunctions(t *testing.T) {
	t.Parallel()
	proc := BuildOpenAPIWeatherServer(t)
	require.NotNil(t, proc)
	err := proc.Start()
	require.NoError(t, err)
	defer proc.Stop()

	integration.WaitForTCPPort(t, proc.Port, integration.ServiceStartupTimeout)

	mcpanyTestServerInfo := integration.StartMCPANYServer(t, "openapi-weather-test")
	defer mcpanyTestServerInfo.CleanupFunc()

	upstreamEndpoint := fmt.Sprintf("http://localhost:%d", proc.Port)
	RegisterOpenAPIWeatherService(t, mcpanyTestServerInfo.RegistrationClient, upstreamEndpoint)

	verifyToolRegistered(t, mcpanyTestServerInfo.HTTPEndpoint, "e2e_openapi_weather.getWeather", "getWeather tool should be registered for openapi service")
}

func TestOpenAPIAuthedHelperFunctions(t *testing.T) {
	t.Parallel()
	proc := BuildOpenAPIAuthedServer(t)
	require.NotNil(t, proc)
	err := proc.Start()
	require.NoError(t, err)
	defer proc.Stop()

	integration.WaitForTCPPort(t, proc.Port, integration.ServiceStartupTimeout)

	mcpanyTestServerInfo := integration.StartMCPANYServer(t, "openapi-authed-test")
	defer mcpanyTestServerInfo.CleanupFunc()

	upstreamEndpoint := fmt.Sprintf("http://localhost:%d", proc.Port)
	RegisterOpenAPIAuthedService(t, mcpanyTestServerInfo.RegistrationClient, upstreamEndpoint)

	verifyToolRegistered(t, mcpanyTestServerInfo.HTTPEndpoint, "e2e_openapi_authed_echo.echo", "echo tool should be registered for openapi authed service")
}

func TestStreamableHTTPHelperFunctions(t *testing.T) {
	t.Parallel()
	proc := BuildStreamableHTTPServer(t)
	require.NotNil(t, proc)
	err := proc.Start()
	require.NoError(t, err)
	defer proc.Stop()

	integration.WaitForTCPPort(t, proc.Port, integration.ServiceStartupTimeout)

	mcpanyTestServerInfo := integration.StartMCPANYServer(t, "streamable-http-test")
	defer mcpanyTestServerInfo.CleanupFunc()

	upstreamEndpoint := fmt.Sprintf("http://localhost:%d/mcp", proc.Port)
	RegisterStreamableHTTPService(t, mcpanyTestServerInfo.RegistrationClient, upstreamEndpoint)

	verifyToolRegistered(t, mcpanyTestServerInfo.HTTPEndpoint, "e2e_everything_server_streamable.time.getTime", "getTime tool should be registered for streamable http service")
}

func TestValidateRegisteredTool(t *testing.T) {
	t.Parallel()
	mcpanyTestServerInfo := integration.StartMCPANYServer(t, "validate-tool-test")
	defer mcpanyTestServerInfo.CleanupFunc()

	// Use a simple stdio service for testing
	RegisterStdioService(t, mcpanyTestServerInfo.RegistrationClient, "")

	expectedTool := &mcp.Tool{
		Name:        "e2e_everything_server_stdio.time.getTime",
		Description: "Returns the current time",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	}
	ValidateRegisteredTool(t, mcpanyTestServerInfo.HTTPEndpoint, expectedTool)
}

func TestVerifyMCPClient(t *testing.T) {
	t.Parallel()
	mcpanyTestServerInfo := integration.StartMCPANYServer(t, "verify-mcp-client-test")
	defer mcpanyTestServerInfo.CleanupFunc()

	// Use a simple stdio service for testing
	RegisterStdioService(t, mcpanyTestServerInfo.RegistrationClient, "")

	VerifyMCPClient(t, mcpanyTestServerInfo.HTTPEndpoint)
}

func verifyToolRegistered(t *testing.T, mcpanyEndpoint, toolName, message string) {
	require.Eventually(t, func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
		defer cancel()

		client := mcp.NewClient(&mcp.Implementation{Name: "e2e-test-client"}, nil)
		transport := &mcp.StreamableClientTransport{
			Endpoint: mcpanyEndpoint,
		}
		session, err := client.Connect(ctx, transport, nil)
		if err != nil {
			return false
		}
		defer session.Close()

		tools, err := session.ListTools(ctx, &mcp.ListToolsParams{})
		if err != nil {
			return false
		}

		for _, tool := range tools.Tools {
			if tool.Name == toolName {
				return true
			}
		}
		return false
	}, integration.TestWaitTimeMedium, 1*time.Second, message)
}
