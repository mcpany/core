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
	"regexp"
	"strconv"
	"strings"
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
	FileRegistration RegistrationMethod = "file"
	// GRPCRegistration uses the RegistrationService via gRPC.
	GRPCRegistration RegistrationMethod = "grpc"
	// JSONRPCRegistration uses the RegistrationService via JSON-RPC.
	JSONRPCRegistration RegistrationMethod = "jsonrpc"
)

var portRegex = regexp.MustCompile(`(?:metricsPort=|Metrics server listening on port |Listening on port port=)(\d+)`)

// E2ETestCase e2ETestCase represents a e2 e test case.
//
// Summary: E2ETestCase represents a e2 e test case.
type E2ETestCase struct {
	Name                         string
	UpstreamServiceType          string
	BuildUpstream                func(t *testing.T) *integration.ManagedProcess
	RegisterUpstream             func(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string)
	ValidateTool                 func(t *testing.T, mcpanyEndpoint string)
	ValidateMiddlewares          func(t *testing.T, mcpanyEndpoint string, upstreamEndpoint string)
	InvokeAIClient               func(t *testing.T, mcpanyEndpoint string)
	InvokeAIClientWithServerInfo func(t *testing.T, serverInfo *integration.MCPANYTestServerInfo)
	RegistrationMethods          []RegistrationMethod
	GenerateUpstreamConfig       func(_ string) string
	StartMCPANYServer            func(t *testing.T, testName string, extraArgs ...string) *integration.MCPANYTestServerInfo
	RegisterUpstreamWithJSONRPC  func(t *testing.T, mcpanyEndpoint, upstreamEndpoint string)
}

// ValidateRegisteredTool validateRegisteredTool validate registered tool.
//
// Summary: ValidateRegisteredTool validate registered tool.
//
// Parameters:
//   - t (*testing.T): The t.
//   - mcpanyEndpoint (string): The mcpany endpoint.
//   - expectedTool (*mcp.Tool): The expected tool.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// RunE2ETest runE2ETest run e2 e test.
//
// Summary: RunE2ETest run e2 e test.
//
// Parameters:
//   - t (*testing.T): The t.
//   - testCase (*E2ETestCase): The test case.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
				configContent := testCase.GenerateUpstreamConfig(fmt.Sprintf("127.0.0.1:%d", upstreamServerProc.Port))
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
				upstreamEndpoint = fmt.Sprintf("127.0.0.1:%d", upstreamServerProc.Port)
			case "websocket":
				upstreamEndpoint = fmt.Sprintf("ws://127.0.0.1:%d/echo", upstreamServerProc.Port)
			case "webrtc":
				upstreamEndpoint = fmt.Sprintf("http://127.0.0.1:%d/signal", upstreamServerProc.Port)
			case "openapi":
				upstreamEndpoint = fmt.Sprintf("http://127.0.0.1:%d", upstreamServerProc.Port)
			case "streamablehttp":
				upstreamEndpoint = fmt.Sprintf("http://127.0.0.1:%d/mcp", upstreamServerProc.Port)
			default:
				upstreamEndpoint = fmt.Sprintf("http://127.0.0.1:%d", upstreamServerProc.Port)
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

// BuildGRPCWeatherServer buildGRPCWeatherServer build grpc weather server.
//
// Summary: BuildGRPCWeatherServer build grpc weather server.
//
// Parameters:
//   - t (*testing.T): The t.
//
// Returns:
//   - *integration.ManagedProcess: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func BuildGRPCWeatherServer(t *testing.T) *integration.ManagedProcess {
	port := 0
	proc := integration.NewManagedProcess(t, "grpc_weather_server", integration.MockBinary(t, "grpc_weather_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

// RegisterGRPCWeatherService registerGRPCWeatherService register grpc weather service.
//
// Summary: RegisterGRPCWeatherService register grpc weather service.
//
// Parameters:
//   - t (*testing.T): The t.
//   - registrationClient (apiv1.RegistrationServiceClient): The registration client.
//   - upstreamEndpoint (string): The upstream endpoint.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func RegisterGRPCWeatherService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_grpc_weather"
	integration.RegisterGRPCService(t, registrationClient, serviceID, upstreamEndpoint, nil)
}

// BuildGRPCAuthedWeatherServer buildGRPCAuthedWeatherServer build grpc authed weather server.
//
// Summary: BuildGRPCAuthedWeatherServer build grpc authed weather server.
//
// Parameters:
//   - t (*testing.T): The t.
//
// Returns:
//   - *integration.ManagedProcess: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func BuildGRPCAuthedWeatherServer(t *testing.T) *integration.ManagedProcess {
	port := 0
	proc := integration.NewManagedProcess(t, "grpc_authed_weather_server", integration.MockBinary(t, "grpc_authed_weather_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

// WaitForPort waitForPort wait for port.
//
// Summary: WaitForPort wait for port.
//
// Parameters:
//   - t (*testing.T): The t.
//   - proc (*integration.ManagedProcess): The proc.
//
// Returns:
//   - int: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func WaitForPort(t *testing.T, proc *integration.ManagedProcess) int {
	t.Helper()
	var port int
	timeout := 10 * time.Second
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for time.Now().Before(deadline) {
		// Try simple parsing
		out := proc.StdoutString()
		lines := strings.Split(out, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if p, err := strconv.Atoi(line); err == nil && p > 0 {
				port = p
				t.Logf("WaitForPort: Found simplified port %d", port)
				return port
			}
		}

		// Try regex on Stdout
		matches := portRegex.FindStringSubmatch(out)
		if len(matches) >= 2 {
			if p, err := strconv.Atoi(matches[1]); err == nil {
				port = p
				t.Logf("WaitForPort: Found regex port %d in Stdout", port)
				return port
			}
		}

		// Try regex on Stderr
		outErr := proc.StderrString()
		matchesErr := portRegex.FindStringSubmatch(outErr)
		if len(matchesErr) >= 2 {
			if p, err := strconv.Atoi(matchesErr[1]); err == nil {
				port = p
				t.Logf("WaitForPort: Found regex port %d in Stderr", port)
				return port
			}
		}

		select {
		case <-ticker.C:
			continue
		default:
		}
	}

	t.Fatalf("Process %s did not output a port within %v. Stdout: %q, Stderr: %q", proc.Cmd().Path, timeout, proc.StdoutString(), proc.StderrString())
	return 0
}

// RegisterGRPCAuthedWeatherService registerGRPCAuthedWeatherService register grpc authed weather service.
//
// Summary: RegisterGRPCAuthedWeatherService register grpc authed weather service.
//
// Parameters:
//   - t (*testing.T): The t.
//   - registrationClient (apiv1.RegistrationServiceClient): The registration client.
//   - upstreamEndpoint (string): The upstream endpoint.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func RegisterGRPCAuthedWeatherService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_grpc_authed_weather"
	secret := configv1.SecretValue_builder{
		PlainText: proto.String("test-bearer-token"),
	}.Build()
	authConfig := configv1.Authentication_builder{
		BearerToken: configv1.BearerTokenAuth_builder{
			Token: secret,
		}.Build(),
	}.Build()
	integration.RegisterGRPCService(t, registrationClient, serviceID, upstreamEndpoint, authConfig)
}

// BuildWebsocketWeatherServer buildWebsocketWeatherServer build websocket weather server.
//
// Summary: BuildWebsocketWeatherServer build websocket weather server.
//
// Parameters:
//   - t (*testing.T): The t.
//
// Returns:
//   - *integration.ManagedProcess: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func BuildWebsocketWeatherServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)
	binaryPath := filepath.Join(t.TempDir(), "weather-server")
	sourcePath := filepath.Join(root, "examples/upstream_service_demo/http/server/weather_server/weather_server.go")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "build", "-o", binaryPath, sourcePath)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	require.NoError(t, err, "Failed to build weather-server")
	proc := integration.NewManagedProcess(t, "websocket_weather_server", binaryPath, []string{fmt.Sprintf("--addr=127.0.0.1:%d", port)}, []string{fmt.Sprintf("HTTP_PORT=%d", port)})
	proc.Port = port
	return proc
}

// RegisterWebsocketWeatherService registerWebsocketWeatherService register websocket weather service.
//
// Summary: RegisterWebsocketWeatherService register websocket weather service.
//
// Parameters:
//   - t (*testing.T): The t.
//   - registrationClient (apiv1.RegistrationServiceClient): The registration client.
//   - upstreamEndpoint (string): The upstream endpoint.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func RegisterWebsocketWeatherService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_websocket_weather"
	integration.RegisterWebsocketService(t, registrationClient, serviceID, upstreamEndpoint, "weather", nil)
}

// BuildWebrtcWeatherServer buildWebrtcWeatherServer build webrtc weather server.
//
// Summary: BuildWebrtcWeatherServer build webrtc weather server.
//
// Parameters:
//   - t (*testing.T): The t.
//
// Returns:
//   - *integration.ManagedProcess: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func BuildWebrtcWeatherServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	proc := integration.NewManagedProcess(t, "webrtc_weather_server", integration.MockBinary(t, "webrtc_weather_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

// RegisterWebrtcWeatherService registerWebrtcWeatherService register webrtc weather service.
//
// Summary: RegisterWebrtcWeatherService register webrtc weather service.
//
// Parameters:
//   - t (*testing.T): The t.
//   - registrationClient (apiv1.RegistrationServiceClient): The registration client.
//   - upstreamEndpoint (string): The upstream endpoint.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func RegisterWebrtcWeatherService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_webrtc_weather"
	integration.RegisterWebrtcService(t, registrationClient, serviceID, upstreamEndpoint, "weather", nil)
}

// BuildStdioServer buildStdioServer build stdio server.
//
// Summary: BuildStdioServer build stdio server.
//
// Parameters:
//   - _ (*testing.T): Unused parameter.
//
// Returns:
//   - *integration.ManagedProcess: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func BuildStdioServer(_ *testing.T) *integration.ManagedProcess {
	return nil
}

// RegisterStdioService registerStdioService register stdio service.
//
// Summary: RegisterStdioService register stdio service.
//
// Parameters:
//   - t (*testing.T): The t.
//   - registrationClient (apiv1.RegistrationServiceClient): The registration client.
//   - _ (string): Unused parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func RegisterStdioService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, _ string) {
	const serviceID = "e2e_everything_server_stdio"
	serviceStdioEndpoint := "npx @modelcontextprotocol/server-everything stdio"
	integration.RegisterStdioMCPService(t, registrationClient, serviceID, serviceStdioEndpoint, true)
}

// BuildStdioDockerServer buildStdioDockerServer build stdio docker server.
//
// Summary: BuildStdioDockerServer build stdio docker server.
//
// Parameters:
//   - _ (*testing.T): Unused parameter.
//
// Returns:
//   - *integration.ManagedProcess: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func BuildStdioDockerServer(_ *testing.T) *integration.ManagedProcess {
	return nil
}

// RegisterStdioDockerService registerStdioDockerService register stdio docker service.
//
// Summary: RegisterStdioDockerService register stdio docker service.
//
// Parameters:
//   - t (*testing.T): The t.
//   - registrationClient (apiv1.RegistrationServiceClient): The registration client.
//   - _ (string): Unused parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func RegisterStdioDockerService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, _ string) {
	const serviceID = "e2e-cowsay-server"
	command := "/app/cowsay_server_bin"
	args := []string{"--mcp-stdio"}
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

// BuildOpenAPIWeatherServer buildOpenAPIWeatherServer build open api weather server.
//
// Summary: BuildOpenAPIWeatherServer build open api weather server.
//
// Parameters:
//   - t (*testing.T): The t.
//
// Returns:
//   - *integration.ManagedProcess: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func BuildOpenAPIWeatherServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	proc := integration.NewManagedProcess(t, "openapi_weather_server", integration.MockBinary(t, "openapi_weather_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

// RegisterOpenAPIWeatherService registerOpenAPIWeatherService register open api weather service.
//
// Summary: RegisterOpenAPIWeatherService register open api weather service.
//
// Parameters:
//   - t (*testing.T): The t.
//   - registrationClient (apiv1.RegistrationServiceClient): The registration client.
//   - upstreamEndpoint (string): The upstream endpoint.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// BuildOpenAPIAuthedServer buildOpenAPIAuthedServer build open api authed server.
//
// Summary: BuildOpenAPIAuthedServer build open api authed server.
//
// Parameters:
//   - t (*testing.T): The t.
//
// Returns:
//   - *integration.ManagedProcess: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func BuildOpenAPIAuthedServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	proc := integration.NewManagedProcess(t, "http_authed_echo_server_openapi", integration.MockBinary(t, "http_authed_echo_server"), []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

// RegisterOpenAPIAuthedService registerOpenAPIAuthedService register open api authed service.
//
// Summary: RegisterOpenAPIAuthedService register open api authed service.
//
// Parameters:
//   - t (*testing.T): The t.
//   - registrationClient (apiv1.RegistrationServiceClient): The registration client.
//   - upstreamEndpoint (string): The upstream endpoint.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
	authConfig := configv1.Authentication_builder{
		ApiKey: configv1.APIKeyAuth_builder{
			ParamName: proto.String("X-Api-Key"),
			Value:     secret,
		}.Build(),
	}.Build()
	integration.RegisterOpenAPIService(t, registrationClient, serviceID, tmpfile.Name(), upstreamEndpoint, authConfig)
}

// BuildStreamableHTTPServer buildStreamableHTTPServer build streamable http server.
//
// Summary: BuildStreamableHTTPServer build streamable http server.
//
// Parameters:
//   - t (*testing.T): The t.
//
// Returns:
//   - *integration.ManagedProcess: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func BuildStreamableHTTPServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	args := []string{"@modelcontextprotocol/server-everything", "streamableHttp"}
	env := []string{fmt.Sprintf("PORT=%d", port)}
	proc := integration.NewManagedProcess(t, "everything_streamable_server", "npx", args, env)
	proc.IgnoreExitStatusOne = true
	proc.Port = port
	return proc
}

// RegisterStreamableHTTPService registerStreamableHTTPService register streamable http service.
//
// Summary: RegisterStreamableHTTPService register streamable http service.
//
// Parameters:
//   - t (*testing.T): The t.
//   - registrationClient (apiv1.RegistrationServiceClient): The registration client.
//   - upstreamEndpoint (string): The upstream endpoint.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func RegisterStreamableHTTPService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_everything_server_streamable"
	integration.RegisterStreamableMCPService(t, registrationClient, serviceID, upstreamEndpoint, true, nil)
}

// VerifyMCPClient verifyMCPClient verify mcp client.
//
// Summary: VerifyMCPClient verify mcp client.
//
// Parameters:
//   - t (*testing.T): The t.
//   - mcpanyEndpoint (string): The mcpany endpoint.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
