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
	"fmt"
	"net/http"
	"testing"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/tests/integration"
	"github.com/stretchr/testify/require"
)

func TestBuildGRPCWeatherServer(t *testing.T) {
	proc := BuildGRPCWeatherServer(t)
	require.NotNil(t, proc)
	err := proc.Start()
	require.NoError(t, err)
	defer proc.Stop()
	integration.WaitForTCPPort(t, proc.Port, integration.ServiceStartupTimeout)
}

func TestRegisterGRPCWeatherService(t *testing.T) {
	testServerInfo := integration.StartMCPANYServer(t, "grpc-weather-test")
	defer testServerInfo.CleanupFunc()

	upstreamEndpoint := "localhost:12345"
	RegisterGRPCWeatherService(t, testServerInfo.RegistrationClient, upstreamEndpoint)
}

func TestBuildGRPCAuthedWeatherServer(t *testing.T) {
	proc := BuildGRPCAuthedWeatherServer(t)
	require.NotNil(t, proc)
	err := proc.Start()
	require.NoError(t, err)
	defer proc.Stop()
	integration.WaitForTCPPort(t, proc.Port, integration.ServiceStartupTimeout)
}

func TestRegisterGRPCAuthedWeatherService(t *testing.T) {
	testServerInfo := integration.StartMCPANYServer(t, "grpc-authed-weather-test")
	defer testServerInfo.CleanupFunc()

	upstreamEndpoint := "localhost:54321"
	RegisterGRPCAuthedWeatherService(t, testServerInfo.RegistrationClient, upstreamEndpoint)
}

func TestBuildWebsocketWeatherServer(t *testing.T) {
	proc := BuildWebsocketWeatherServer(t)
	require.NotNil(t, proc)
	err := proc.Start()
	require.NoError(t, err)
	defer proc.Stop()
	integration.WaitForTCPPort(t, proc.Port, integration.ServiceStartupTimeout)
}

func TestRegisterWebsocketWeatherService(t *testing.T) {
	testServerInfo := integration.StartMCPANYServer(t, "websocket-weather-test")
	defer testServerInfo.CleanupFunc()

	upstreamEndpoint := "ws://localhost:12345/echo"
	RegisterWebsocketWeatherService(t, testServerInfo.RegistrationClient, upstreamEndpoint)
}

func TestBuildWebrtcWeatherServer(t *testing.T) {
	proc := BuildWebrtcWeatherServer(t)
	require.NotNil(t, proc)
	err := proc.Start()
	require.NoError(t, err)
	defer proc.Stop()
	integration.WaitForTCPPort(t, proc.Port, integration.ServiceStartupTimeout)
}

func TestRegisterWebrtcWeatherService(t *testing.T) {
	testServerInfo := integration.StartMCPANYServer(t, "webrtc-weather-test")
	defer testServerInfo.CleanupFunc()

	upstreamEndpoint := "http://localhost:12345/signal"
	RegisterWebrtcWeatherService(t, testServerInfo.RegistrationClient, upstreamEndpoint)
}

func TestBuildStdioServer(t *testing.T) {
	proc := BuildStdioServer(t)
	require.Nil(t, proc)
}

func TestRegisterStdioService(t *testing.T) {
	testServerInfo := integration.StartMCPANYServer(t, "stdio-test")
	defer testServerInfo.CleanupFunc()
	RegisterStdioService(t, testServerInfo.RegistrationClient, "")
}

func TestBuildStdioDockerServer(t *testing.T) {
	proc := BuildStdioDockerServer(t)
	require.Nil(t, proc)
}

func TestRegisterStdioDockerService(t *testing.T) {
	testServerInfo := integration.StartMCPANYServer(t, "stdio-docker-test")
	defer testServerInfo.CleanupFunc()
	RegisterStdioDockerService(t, testServerInfo.RegistrationClient, "")
}

func TestBuildOpenAPIWeatherServer(t *testing.T) {
	proc := BuildOpenAPIWeatherServer(t)
	require.NotNil(t, proc)
	err := proc.Start()
	require.NoError(t, err)
	defer proc.Stop()
	integration.WaitForTCPPort(t, proc.Port, integration.ServiceStartupTimeout)
}

func TestRegisterOpenAPIWeatherService(t *testing.T) {
	proc := BuildOpenAPIWeatherServer(t)
	require.NotNil(t, proc)
	err := proc.Start()
	require.NoError(t, err)
	defer proc.Stop()
	integration.WaitForTCPPort(t, proc.Port, integration.ServiceStartupTimeout)

	testServerInfo := integration.StartMCPANYServer(t, "openapi-weather-test")
	defer testServerInfo.CleanupFunc()

	upstreamEndpoint := fmt.Sprintf("http://localhost:%d", proc.Port)
	RegisterOpenAPIWeatherService(t, testServerInfo.RegistrationClient, upstreamEndpoint)
}

func TestBuildOpenAPIAuthedServer(t *testing.T) {
	proc := BuildOpenAPIAuthedServer(t)
	require.NotNil(t, proc)
	err := proc.Start()
	require.NoError(t, err)
	defer proc.Stop()
	integration.WaitForTCPPort(t, proc.Port, integration.ServiceStartupTimeout)
}

func TestRegisterOpenAPIAuthedService(t *testing.T) {
	testServerInfo := integration.StartMCPANYServer(t, "openapi-authed-test")
	defer testServerInfo.CleanupFunc()

	upstreamEndpoint := "http://localhost:12345"
	RegisterOpenAPIAuthedService(t, testServerInfo.RegistrationClient, upstreamEndpoint)
}

func TestBuildStreamableHTTPServer(t *testing.T) {
	proc := BuildStreamableHTTPServer(t)
	require.NotNil(t, proc)
	err := proc.Start()
	require.NoError(t, err)
	defer proc.Stop()
	integration.WaitForTCPPort(t, proc.Port, integration.ServiceStartupTimeout)
}

func TestRegisterStreamableHTTPService(t *testing.T) {
	testServerInfo := integration.StartMCPANYServer(t, "streamable-http-test")
	defer testServerInfo.CleanupFunc()
	upstreamEndpoint := "http://localhost:12345/mcp"
	RegisterStreamableHTTPService(t, testServerInfo.RegistrationClient, upstreamEndpoint)
}

func TestVerifyMCPClient(t *testing.T) {
	mcpanyTestServerInfo := integration.StartMCPANYServer(t, "verify-mcp-client-test")
	defer mcpanyTestServerInfo.CleanupFunc()
	VerifyMCPClient(t, mcpanyTestServerInfo.HTTPEndpoint)
}

func TestE2ETestCase(t *testing.T) {
	mcpanyTestServerInfo := integration.StartMCPANYServer(t, "e2e-test-case")
	defer mcpanyTestServerInfo.CleanupFunc()

	root, err := integration.GetProjectRoot()
	require.NoError(t, err)

	testCase := &E2ETestCase{
		Name:                "test-case",
		UpstreamServiceType: "http",
		BuildUpstream: func(t *testing.T) *integration.ManagedProcess {
			port := integration.FindFreePort(t)
			proc := integration.NewManagedProcess(t, "http-echo-server", root+"/build/test/bin/http_echo_server", []string{fmt.Sprintf("--port=%d", port)}, nil)
			proc.Port = port
			return proc
		},
		RegisterUpstream: func(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
			const serviceID = "e2e-http-echo"
			integration.RegisterHTTPService(t, registrationClient, serviceID, upstreamEndpoint, "echo", "/echo", http.MethodPost, nil)
		},
		ValidateTool: func(t *testing.T, mcpanyEndpoint string) {
		},
		ValidateMiddlewares: func(t *testing.T, mcpanyEndpoint string, upstreamEndpoint string) {
		},
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
		},
		RegistrationMethods: []RegistrationMethod{GRPCRegistration},
		GenerateUpstreamConfig: func(upstreamEndpoint string) string {
			return ""
		},
		StartMCPANYServer: func(t *testing.T, testName string, extraArgs ...string) *integration.MCPANYTestServerInfo {
			return mcpanyTestServerInfo
		},
		RegisterUpstreamWithJSONRPC: func(t *testing.T, mcpanyEndpoint, upstreamEndpoint string) {
			integration.RegisterHTTPServiceWithJSONRPC(t, mcpanyEndpoint, "e2e-http-echo", upstreamEndpoint, "echo", "/echo", http.MethodPost, nil)
		},
	}
	RunE2ETest(t, testCase)

	testCase.RegistrationMethods = []RegistrationMethod{FileRegistration}
	testCase.GenerateUpstreamConfig = func(upstreamEndpoint string) string {
		return fmt.Sprintf(`
services:
- id: e2e-http-echo
  http:
    url: %s
`, upstreamEndpoint)
	}

	testCase.StartMCPANYServer = func(t *testing.T, testName string, extraArgs ...string) *integration.MCPANYTestServerInfo {
		configContent := testCase.GenerateUpstreamConfig(fmt.Sprintf("localhost:%d", 1234))
		return integration.StartMCPANYServerWithConfig(t, testName, configContent)
	}
	RunE2ETest(t, testCase)

	testCase.RegistrationMethods = []RegistrationMethod{JSONRPCRegistration}

	testCase.StartMCPANYServer = func(t *testing.T, testName string, extraArgs ...string) *integration.MCPANYTestServerInfo {
		return integration.StartMCPANYServer(t, testName)
	}

	RunE2ETest(t, testCase)
}
