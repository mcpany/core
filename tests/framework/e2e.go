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
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"

	"github.com/mcpany/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

type RegistrationMethod string

const (
	FileRegistration    RegistrationMethod = "file"
	GRPCRegistration    RegistrationMethod = "grpc"
	JSONRPCRegistration RegistrationMethod = "jsonrpc"
)

type E2ETestCase struct {
	Name                        string
	UpstreamServiceType         string
	BuildUpstream               func(t *testing.T) *integration.ManagedProcess
	RegisterUpstream            func(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string)
	ValidateTool                func(t *testing.T, mcpanyEndpoint string)
	ValidateMiddlewares         func(t *testing.T, mcpanyEndpoint string, upstreamEndpoint string)
	InvokeAIClient              func(t *testing.T, mcpanyEndpoint string)
	RegistrationMethods         []RegistrationMethod
	GenerateUpstreamConfig      func(upstreamEndpoint string) string
	StartMCPANYServer           func(t *testing.T, testName string, extraArgs ...string) *integration.MCPANYTestServerInfo
	RegisterUpstreamWithJSONRPC func(t *testing.T, mcpanyEndpoint, upstreamEndpoint string)
}

func ValidateRegisteredTool(t *testing.T, mcpanyEndpoint string, expectedTool *mcp.Tool) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	client := mcp.NewClient(&mcp.Implementation{Name: "e2e-test-client"}, nil)

	transport := &mcp.StreamableClientTransport{
		Endpoint: mcpanyEndpoint,
	}

	session, err := client.Connect(ctx, transport, nil)
	require.NoError(t, err)
	defer session.Close()

	tools, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)

	var foundTool *mcp.Tool
	for _, tool := range tools.Tools {
		if tool.Name == expectedTool.Name {
			foundTool = tool
			break
		}
	}

	require.NotNil(t, foundTool, "tool %q not found", expectedTool.Name)
	require.Equal(t, expectedTool.Description, foundTool.Description)
	require.Equal(t, expectedTool.InputSchema, foundTool.InputSchema)
}

func RunE2ETest(t *testing.T, testCase *E2ETestCase) {
	for _, method := range testCase.RegistrationMethods {
		method := method
		t.Run(fmt.Sprintf("%s_%s", testCase.Name, method), func(t *testing.T) {
			_, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			t.Logf("INFO: Starting E2E Test Scenario for %s with %s registration...", testCase.Name, method)
			t.Parallel()

			// --- 1. Start Upstream Service ---
			var upstreamServerProc *integration.ManagedProcess
			if testCase.BuildUpstream != nil {
				upstreamServerProc = testCase.BuildUpstream(t)
				if upstreamServerProc != nil {
					err := upstreamServerProc.Start()
					require.NoError(t, err, "Failed to start upstream server")
					t.Cleanup(upstreamServerProc.Stop)
					if upstreamServerProc.Port != 0 {
						integration.WaitForTCPPort(t, upstreamServerProc.Port, integration.ServiceStartupTimeout)
					}
				}
			}

			var mcpanyTestServerInfo *integration.MCPANYTestServerInfo
			switch {
			case testCase.StartMCPANYServer != nil:
				mcpanyTestServerInfo = testCase.StartMCPANYServer(t, testCase.Name)
			case method == FileRegistration:
				configContent := testCase.GenerateUpstreamConfig(fmt.Sprintf("localhost:%d", upstreamServerProc.Port))
				mcpanyTestServerInfo = integration.StartMCPANYServerWithConfig(t, testCase.Name, configContent)
			default:
				mcpanyTestServerInfo = integration.StartMCPANYServer(t, testCase.Name)
			}
			defer mcpanyTestServerInfo.CleanupFunc()

			// Add a small delay to ensure the server is ready to accept registrations
			time.Sleep(1 * time.Second)

			// --- 3. Register Upstream Service with MCPANY ---
			var upstreamEndpoint string
			if testCase.UpstreamServiceType == "stdio" {
				upstreamEndpoint = ""
			} else if testCase.UpstreamServiceType == "grpc" {
				upstreamEndpoint = fmt.Sprintf("localhost:%d", upstreamServerProc.Port)
			} else if testCase.UpstreamServiceType == "websocket" {
				upstreamEndpoint = fmt.Sprintf("ws://localhost:%d/echo", upstreamServerProc.Port)
			} else if testCase.UpstreamServiceType == "webrtc" {
				upstreamEndpoint = fmt.Sprintf("http://localhost:%d/signal", upstreamServerProc.Port)
			} else if testCase.UpstreamServiceType == "openapi" {
				upstreamEndpoint = fmt.Sprintf("http://localhost:%d", upstreamServerProc.Port)
			} else if testCase.UpstreamServiceType == "streamablehttp" {
				upstreamEndpoint = fmt.Sprintf("http://localhost:%d/mcp", upstreamServerProc.Port)
			} else {
				upstreamEndpoint = fmt.Sprintf("http://localhost:%d", upstreamServerProc.Port)
			}
			t.Logf("INFO: Registering upstream service with MCPANY at endpoint %s...", upstreamEndpoint)
			switch method {
			case GRPCRegistration:
				testCase.RegisterUpstream(t, mcpanyTestServerInfo.RegistrationClient, upstreamEndpoint)
			case JSONRPCRegistration:
				testCase.RegisterUpstreamWithJSONRPC(t, mcpanyTestServerInfo.HTTPEndpoint, upstreamEndpoint)
			}
			t.Logf("INFO: Upstream service registered.")

			// --- 4. Validate Registered Tool ---
			if testCase.ValidateTool != nil {
				t.Logf("INFO: Validating registered tool...")
				testCase.ValidateTool(t, mcpanyTestServerInfo.HTTPEndpoint)
				t.Logf("INFO: Tool validated.")
			}

			// --- 5. Validate Middlewares ---
			if testCase.ValidateMiddlewares != nil {
				t.Logf("INFO: Validating middlewares...")
				testCase.ValidateMiddlewares(t, mcpanyTestServerInfo.HTTPEndpoint, upstreamEndpoint)
				t.Logf("INFO: Middlewares validated.")
			}

			// --- 5. Invoke AI Client ---
			testCase.InvokeAIClient(t, mcpanyTestServerInfo.HTTPEndpoint)

			t.Logf("INFO: E2E Test Scenario for %s Completed Successfully!", testCase.Name)
		})
	}
}

func BuildGRPCWeatherServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)
	proc := integration.NewManagedProcess(t, "grpc_weather_server", filepath.Join(root, "build/test/bin/grpc_weather_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

func RegisterGRPCWeatherService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_grpc_weather"
	integration.RegisterGRPCService(t, registrationClient, serviceID, upstreamEndpoint, nil)
}

func BuildGRPCAuthedWeatherServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)
	proc := integration.NewManagedProcess(t, "grpc_authed_weather_server", filepath.Join(root, "build/test/bin/grpc_authed_weather_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

func RegisterGRPCAuthedWeatherService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_grpc_authed_weather"
	secret := configv1.SecretValue_builder{
		PlainText: proto.String("test-bearer-token"),
	}.Build()
	authConfig := configv1.UpstreamAuthentication_builder{
		BearerToken: configv1.UpstreamBearerTokenAuth_builder{
			Token: secret,
		}.Build(),
	}.Build()
	integration.RegisterGRPCService(t, registrationClient, serviceID, upstreamEndpoint, authConfig)
}

func BuildWebsocketWeatherServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)
	proc := integration.NewManagedProcess(t, "websocket_weather_server", filepath.Join(root, "build/examples/bin/weather-server"), []string{fmt.Sprintf("--addr=localhost:%d", port)}, []string{fmt.Sprintf("HTTP_PORT=%d", port)})
	proc.Port = port
	return proc
}

func RegisterWebsocketWeatherService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_websocket_weather"
	integration.RegisterWebsocketService(t, registrationClient, serviceID, upstreamEndpoint, "weather", nil)
}

func BuildWebrtcWeatherServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)
	proc := integration.NewManagedProcess(t, "webrtc_weather_server", filepath.Join(root, "build/test/bin/webrtc_weather_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

func RegisterWebrtcWeatherService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_webrtc_weather"
	integration.RegisterWebrtcService(t, registrationClient, serviceID, upstreamEndpoint, "weather", nil)
}

func BuildStdioServer(t *testing.T) *integration.ManagedProcess {
	return nil
}

func RegisterStdioService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_everything_server_stdio"
	serviceStdioEndpoint := "npx @modelcontextprotocol/server-everything stdio"
	integration.RegisterStdioMCPService(t, registrationClient, serviceID, serviceStdioEndpoint, true)
}

func BuildStdioDockerServer(t *testing.T) *integration.ManagedProcess {
	return nil
}

func RegisterStdioDockerService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e-cowsay-server"
	command := "python3"
	args := []string{"-u", "main.py", "--mcp-stdio"}
	setupCommands := []string{"pip install -q cowsay"}
	integration.RegisterStdioServiceWithSetup(
		t,
		registrationClient,
		serviceID,
		command,
		true,
		"/work/tests/integration/cmd/mocks/python_cowsay_server", // working directory
		"python:3.11-slim", // No explicit container image
		setupCommands,
		args...,
	)
}

func BuildOpenAPIWeatherServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)
	proc := integration.NewManagedProcess(t, "openapi_weather_server", filepath.Join(root, "build/test/bin/openapi_weather_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

func RegisterOpenAPIWeatherService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_openapi_weather"
	openapiSpecEndpoint := fmt.Sprintf("%s/openapi.json", upstreamEndpoint)
	resp, err := http.Get(openapiSpecEndpoint) //nolint:gosec
	require.NoError(t, err, "Failed to fetch OpenAPI spec from server")
	defer resp.Body.Close()
	specContent, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read OpenAPI spec content")
	tmpfile, err := os.CreateTemp("", "openapi-*.json")
	require.NoError(t, err, "Failed to create temp file for OpenAPI spec")
	defer os.Remove(tmpfile.Name())
	_, err = tmpfile.Write(specContent)
	require.NoError(t, err, "Failed to write spec to temp file")
	err = tmpfile.Close()
	require.NoError(t, err, "Failed to close temp file")
	integration.RegisterOpenAPIService(t, registrationClient, serviceID, tmpfile.Name(), upstreamEndpoint, nil)
}

func BuildOpenAPIAuthedServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)
	proc := integration.NewManagedProcess(t, "http_authed_echo_server_openapi", filepath.Join(root, "build/test/bin/http_authed_echo_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

func RegisterOpenAPIAuthedService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_openapi_authed_echo"
	openapiSpec := fmt.Sprintf(`
openapi: 3.0.0
info:
  title: Authenticated Echo Service
  version: 1.0.0
servers:
  - url: %s
paths:
  /echo:
    post:
      operationId: echo
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                message:
                  type: string
      responses:
        '200':
          description: OK
`, upstreamEndpoint)
	tmpfile, err := os.CreateTemp("", "openapi-auth-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	_, err = tmpfile.WriteString(openapiSpec)
	require.NoError(t, err)
	err = tmpfile.Close()
	require.NoError(t, err)
	secret := configv1.SecretValue_builder{
		PlainText: proto.String("test-api-key"),
	}.Build()
	authConfig := configv1.UpstreamAuthentication_builder{
		ApiKey: configv1.UpstreamAPIKeyAuth_builder{
			HeaderName: proto.String("X-Api-Key"),
			ApiKey:     secret,
		}.Build(),
	}.Build()
	integration.RegisterOpenAPIService(t, registrationClient, serviceID, tmpfile.Name(), upstreamEndpoint, authConfig)
}

func BuildStreamableHTTPServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	args := []string{"@modelcontextprotocol/server-everything", "streamableHttp"}
	env := []string{fmt.Sprintf("PORT=%d", port)}
	proc := integration.NewManagedProcess(t, "everything_streamable_server", "npx", args, env)
	proc.IgnoreExitStatusOne = true
	proc.Port = port
	return proc
}

func RegisterStreamableHTTPService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_everything_server_streamable"
	integration.RegisterStreamableMCPService(t, registrationClient, serviceID, upstreamEndpoint, true, nil)
}

func VerifyMCPClient(t *testing.T, mcpanyEndpoint string) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	for _, tool := range listToolsResult.Tools {
		t.Logf("Discovered tool from MCPANY: %s", tool.Name)
	}
}
