/*
 * Copyright 2025 Author(s) of MCP-XY
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

	apiv1 "github.com/mcpxy/core/proto/api/v1"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"google.golang.org/protobuf/proto"

	"github.com/mcpxy/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

type RegistrationMethod string

const (
	FileRegistration   RegistrationMethod = "file"
	GRPCRegistration   RegistrationMethod = "grpc"
	JSONRPCRegistration RegistrationMethod = "jsonrpc"
)

type E2ETestCase struct {
	Name                   string
	UpstreamServiceType    string
	BuildUpstream          func(t *testing.T) *integration.ManagedProcess
	RegisterUpstream       func(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string)
	ValidateTool           func(t *testing.T, mcpxyEndpoint string)
	InvokeAIClient         func(t *testing.T, mcpxyEndpoint string)
	RegistrationMethods    []RegistrationMethod
	GenerateUpstreamConfig func(upstreamEndpoint string) string
	StartMCPXYServer       func(t *testing.T, testName string, extraArgs ...string) *integration.MCPXYTestServerInfo
}

func ValidateRegisteredTool(t *testing.T, mcpxyEndpoint string, expectedTool *mcp.Tool) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	client := mcp.NewClient(&mcp.Implementation{Name: "e2e-test-client"}, nil)

	transport := &mcp.StreamableClientTransport{
		Endpoint: mcpxyEndpoint,
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

			var mcpxyTestServerInfo *integration.MCPXYTestServerInfo
			if testCase.StartMCPXYServer != nil {
				mcpxyTestServerInfo = testCase.StartMCPXYServer(t, testCase.Name)
			} else if method == FileRegistration {
				configContent := testCase.GenerateUpstreamConfig(fmt.Sprintf("localhost:%d", upstreamServerProc.Port))
				mcpxyTestServerInfo = integration.StartMCPXYServerWithConfig(t, testCase.Name, configContent)
			} else {
				mcpxyTestServerInfo = integration.StartMCPXYServer(t, testCase.Name)
			}
			defer mcpxyTestServerInfo.CleanupFunc()

			// Add a small delay to ensure the server is ready to accept registrations
			time.Sleep(1 * time.Second)

			// --- 3. Register Upstream Service with MCPXY ---
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
			t.Logf("INFO: Registering upstream service with MCPXY at endpoint %s...", upstreamEndpoint)
			if method == GRPCRegistration {
				testCase.RegisterUpstream(t, mcpxyTestServerInfo.RegistrationClient, upstreamEndpoint)
			}
			//TODO: JSONRPC registration
			t.Logf("INFO: Upstream service registered.")

			// --- 4. Validate Registered Tool ---
			if testCase.ValidateTool != nil {
				t.Logf("INFO: Validating registered tool...")
				testCase.ValidateTool(t, mcpxyTestServerInfo.HTTPEndpoint)
				t.Logf("INFO: Tool validated.")
			}

			// --- 5. Invoke AI Client ---
			testCase.InvokeAIClient(t, mcpxyTestServerInfo.HTTPEndpoint)

			t.Logf("INFO: E2E Test Scenario for %s Completed Successfully!", testCase.Name)
		})
	}
}

func BuildGRPCServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)
	proc := integration.NewManagedProcess(t, "grpc_calculator_server", filepath.Join(root, "build/test/bin/grpc_calculator_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

func RegisterGRPCService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_grpc_calculator"
	integration.RegisterGRPCService(t, registrationClient, serviceID, upstreamEndpoint, nil)
}

func BuildGRPCAuthedServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)
	proc := integration.NewManagedProcess(t, "grpc_authed_calculator_server", filepath.Join(root, "build/test/bin/grpc_authed_calculator_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

func RegisterGRPCAuthedService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_grpc_authed_calculator"
	authConfig := configv1.UpstreamAuthentication_builder{
		BearerToken: configv1.UpstreamBearerTokenAuth_builder{
			Token: proto.String("test-bearer-token"),
		}.Build(),
	}.Build()
	integration.RegisterGRPCService(t, registrationClient, serviceID, upstreamEndpoint, authConfig)
}

func BuildWebsocketServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)
	proc := integration.NewManagedProcess(t, "websocket_echo_server", filepath.Join(root, "build/test/bin/websocket_echo_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

func RegisterWebsocketService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_websocket_echo"
	integration.RegisterWebsocketService(t, registrationClient, serviceID, upstreamEndpoint, "echo", nil)
}

func BuildWebrtcServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)
	proc := integration.NewManagedProcess(t, "webrtc_echo_server", filepath.Join(root, "build/test/bin/webrtc_echo_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

func RegisterWebrtcService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_webrtc_echo"
	integration.RegisterWebrtcService(t, registrationClient, serviceID, upstreamEndpoint, "echo", nil)
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

func BuildOpenAPIServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)
	proc := integration.NewManagedProcess(t, "openapi_calculator_server", filepath.Join(root, "build/test/bin/openapi_calculator_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

func RegisterOpenAPIService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_openapi_calculator"
	openapiSpecEndpoint := fmt.Sprintf("%s/openapi.json", upstreamEndpoint)
	resp, err := http.Get(openapiSpecEndpoint)
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
	authConfig := configv1.UpstreamAuthentication_builder{
		ApiKey: configv1.UpstreamAPIKeyAuth_builder{
			HeaderName: proto.String("X-Api-Key"),
			ApiKey:     proto.String("test-api-key"),
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

func VerifyMCPClient(t *testing.T, mcpxyEndpoint string) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxyEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	for _, tool := range listToolsResult.Tools {
		t.Logf("Discovered tool from MCPXY: %s", tool.Name)
	}
}
