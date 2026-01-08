// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

// RegistrationMethod defines the method used to register an upstream service.
type RegistrationMethod string

const (
	// FileRegistration uses a configuration file for registration.
	FileRegistration    RegistrationMethod = "file"
	// GRPCRegistration uses the RegistrationService via gRPC.
	GRPCRegistration    RegistrationMethod = "grpc"
	// JSONRPCRegistration uses the RegistrationService via JSON-RPC.
	JSONRPCRegistration RegistrationMethod = "jsonrpc"
)

// E2ETestCase defines the structure for an end-to-end test case.
type E2ETestCase struct {
	Name                        string
	UpstreamServiceType         string
	BuildUpstream               func(t *testing.T) *integration.ManagedProcess
	RegisterUpstream            func(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string)
	ValidateTool                func(t *testing.T, mcpanyEndpoint string)
	ValidateMiddlewares         func(t *testing.T, mcpanyEndpoint string, upstreamEndpoint string)
	InvokeAIClient              func(t *testing.T, mcpanyEndpoint string)
	InvokeAIClientWithServerInfo func(t *testing.T, serverInfo *integration.MCPANYTestServerInfo)
	RegistrationMethods         []RegistrationMethod
	GenerateUpstreamConfig      func(_ string) string
	StartMCPANYServer           func(t *testing.T, testName string, extraArgs ...string) *integration.MCPANYTestServerInfo
	RegisterUpstreamWithJSONRPC func(t *testing.T, mcpanyEndpoint, upstreamEndpoint string)
}

// ValidateRegisteredTool validates that the expected tool is registered.
func ValidateRegisteredTool(t *testing.T, mcpanyEndpoint string, expectedTool *mcp.Tool) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	client := mcp.NewClient(&mcp.Implementation{Name: "e2e-test-client"}, nil)

	transport := &mcp.StreamableClientTransport{
		Endpoint: mcpanyEndpoint,
	}

	session, err := client.Connect(ctx, transport, nil)
	require.NoError(t, err)
	defer func() { _ = session.Close() }()

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

// RunE2ETest runs an end-to-end test case.
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
			t.Cleanup(mcpanyTestServerInfo.CleanupFunc)

			// Add a small delay to ensure the server is ready to accept registrations
			time.Sleep(1 * time.Second)

			// --- 3. Register Upstream Service with MCPANY ---
			var upstreamEndpoint string
			switch testCase.UpstreamServiceType {
			case "stdio":
				upstreamEndpoint = ""
			case "grpc":
				upstreamEndpoint = fmt.Sprintf("localhost:%d", upstreamServerProc.Port)
			case "websocket":
				upstreamEndpoint = fmt.Sprintf("ws://localhost:%d/echo", upstreamServerProc.Port)
			case "webrtc":
				upstreamEndpoint = fmt.Sprintf("http://localhost:%d/signal", upstreamServerProc.Port)
			case "openapi":
				upstreamEndpoint = fmt.Sprintf("http://localhost:%d", upstreamServerProc.Port)
			case "streamablehttp":
				upstreamEndpoint = fmt.Sprintf("http://localhost:%d/mcp", upstreamServerProc.Port)
			default:
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
			if testCase.InvokeAIClientWithServerInfo != nil {
				testCase.InvokeAIClientWithServerInfo(t, mcpanyTestServerInfo)
			} else if testCase.InvokeAIClient != nil {
				testCase.InvokeAIClient(t, mcpanyTestServerInfo.HTTPEndpoint)
			}

			t.Logf("INFO: E2E Test Scenario for %s Completed Successfully!", testCase.Name)
		})
	}
}

// BuildGRPCWeatherServer builds the gRPC weather server.
func BuildGRPCWeatherServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)
	proc := integration.NewManagedProcess(t, "grpc_weather_server", filepath.Join(root, "../build/test/bin/grpc_weather_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

// RegisterGRPCWeatherService registers the gRPC weather service.
func RegisterGRPCWeatherService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_grpc_weather"
	integration.RegisterGRPCService(t, registrationClient, serviceID, upstreamEndpoint, nil)
}

// BuildGRPCAuthedWeatherServer builds the authenticated gRPC weather server.
func BuildGRPCAuthedWeatherServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)
	proc := integration.NewManagedProcess(t, "grpc_authed_weather_server", filepath.Join(root, "../build/test/bin/grpc_authed_weather_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

// RegisterGRPCAuthedWeatherService registers the authenticated gRPC weather service.
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

// BuildWebsocketWeatherServer builds the websocket weather server.
func BuildWebsocketWeatherServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)
	binaryPath := filepath.Join(root, "../build/examples/bin/weather-server")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Logf("Binary not found at %s, attempting to build...", binaryPath)
		sourcePath := filepath.Join(root, "examples/upstream_service_demo/http/server/weather_server/weather_server.go")
		cmd := exec.Command("go", "build", "-o", binaryPath, sourcePath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		require.NoError(t, err, "Failed to build weather-server")
	}
	proc := integration.NewManagedProcess(t, "websocket_weather_server", binaryPath, []string{fmt.Sprintf("--addr=localhost:%d", port)}, []string{fmt.Sprintf("HTTP_PORT=%d", port)})
	proc.Port = port
	return proc
}

// RegisterWebsocketWeatherService registers the websocket weather service.
func RegisterWebsocketWeatherService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_websocket_weather"
	integration.RegisterWebsocketService(t, registrationClient, serviceID, upstreamEndpoint, "weather", nil)
}

// BuildWebrtcWeatherServer builds the webrtc weather server.
func BuildWebrtcWeatherServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)
	proc := integration.NewManagedProcess(t, "webrtc_weather_server", filepath.Join(root, "../build/test/bin/webrtc_weather_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

// RegisterWebrtcWeatherService registers the webrtc weather service.
func RegisterWebrtcWeatherService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_webrtc_weather"
	integration.RegisterWebrtcService(t, registrationClient, serviceID, upstreamEndpoint, "weather", nil)
}

// BuildStdioServer builds the stdio server (nop).
func BuildStdioServer(_ *testing.T) *integration.ManagedProcess {
	return nil
}

// RegisterStdioService registers the stdio service.
func RegisterStdioService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, _ string) {
	const serviceID = "e2e_everything_server_stdio"
	serviceStdioEndpoint := "npx @modelcontextprotocol/server-everything stdio"
	integration.RegisterStdioMCPService(t, registrationClient, serviceID, serviceStdioEndpoint, true)
}

// BuildStdioDockerServer builds the stdio docker server (nop).
func BuildStdioDockerServer(_ *testing.T) *integration.ManagedProcess {
	return nil
}

// RegisterStdioDockerService registers the stdio docker service.
func RegisterStdioDockerService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, _ string) {
	const serviceID = "e2e-cowsay-server"
	command := "python3"
	args := []string{"-u", "main.py", "--mcp-stdio"}
	// setupCommands := []string{"pip install -q cowsay"}
	integration.RegisterStdioServiceWithSetup(
		t,
		registrationClient,
		serviceID,
		command,
		true,
		"/app", // working directory
		"mcpany/e2e-cowsay-server:latest",
		nil, // No setup commands needed (pre-built image)
		nil,
		args...,
	)
}

// BuildOpenAPIWeatherServer builds the openapi weather server.
func BuildOpenAPIWeatherServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)
	proc := integration.NewManagedProcess(t, "openapi_weather_server", filepath.Join(root, "../build/test/bin/openapi_weather_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

// RegisterOpenAPIWeatherService registers the openapi weather service.
func RegisterOpenAPIWeatherService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_openapi_weather"
	openapiSpecEndpoint := fmt.Sprintf("%s/openapi.json", upstreamEndpoint)
	req, err := http.NewRequestWithContext(context.Background(), "GET", openapiSpecEndpoint, nil)
	require.NoError(t, err, "Failed to create request for OpenAPI spec")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "Failed to fetch OpenAPI spec from server")
	defer func() { _ = resp.Body.Close() }()
	specContent, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read OpenAPI spec content")
	tmpfile, err := os.CreateTemp("", "openapi-*.json")
	require.NoError(t, err, "Failed to create temp file for OpenAPI spec")
	defer func() { _ = os.Remove(tmpfile.Name()) }()
	_, err = tmpfile.Write(specContent)
	require.NoError(t, err, "Failed to write spec to temp file")
	err = tmpfile.Close()
	require.NoError(t, err, "Failed to close temp file")
	integration.RegisterOpenAPIService(t, registrationClient, serviceID, tmpfile.Name(), upstreamEndpoint, nil)
}

// BuildOpenAPIAuthedServer builds the openapi authenticated server.
func BuildOpenAPIAuthedServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)
	proc := integration.NewManagedProcess(t, "http_authed_echo_server_openapi", filepath.Join(root, "../build/test/bin/http_authed_echo_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

// RegisterOpenAPIAuthedService registers the openapi authenticated service.
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
	defer func() { _ = os.Remove(tmpfile.Name()) }()
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

// BuildStreamableHTTPServer builds the streamable http server.
func BuildStreamableHTTPServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	args := []string{"@modelcontextprotocol/server-everything", "streamableHttp"}
	env := []string{fmt.Sprintf("PORT=%d", port)}
	proc := integration.NewManagedProcess(t, "everything_streamable_server", "npx", args, env)
	proc.IgnoreExitStatusOne = true
	proc.Port = port
	return proc
}

// RegisterStreamableHTTPService registers the streamable http service.
func RegisterStreamableHTTPService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_everything_server_streamable"
	integration.RegisterStreamableMCPService(t, registrationClient, serviceID, upstreamEndpoint, true, nil)
}

// VerifyMCPClient verifies the MCP client.
func VerifyMCPClient(t *testing.T, mcpanyEndpoint string) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
	require.NoError(t, err)
	defer func() { _ = cs.Close() }()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	for _, tool := range listToolsResult.Tools {
		t.Logf("Discovered tool from MCPANY: %s", tool.Name)
	}
}
