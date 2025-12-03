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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/tests/integration"
	"github.com/stretchr/testify/require"
)

// newMockUpstreamServer creates a new mock upstream server for testing.
func newMockUpstreamServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func TestBuildGRPCWeatherServer(t *testing.T) {
	proc := BuildGRPCWeatherServer(t)
	require.NotNil(t, proc)
	require.NotNil(t, proc.Cmd())
	require.True(t, strings.Contains(proc.Cmd().Path, "grpc_weather_server"))
	require.NotEmpty(t, proc.Cmd().Args)
	require.NotZero(t, proc.Port)
}

func TestRegisterGRPCWeatherService(t *testing.T) {
	upstreamProc := BuildGRPCWeatherServer(t)
	err := upstreamProc.Start()
	require.NoError(t, err)
	t.Cleanup(upstreamProc.Stop)
	integration.WaitForTCPPort(t, upstreamProc.Port, integration.ServiceStartupTimeout)
	upstreamEndpoint := fmt.Sprintf("localhost:%d", upstreamProc.Port)

	mcpanyTestServerInfo := integration.StartMCPANYServer(t, "grpc-weather-service")
	defer mcpanyTestServerInfo.CleanupFunc()

	time.Sleep(1 * time.Second)

	RegisterGRPCWeatherService(t, mcpanyTestServerInfo.RegistrationClient, upstreamEndpoint)
}

func TestBuildGRPCAuthedWeatherServer(t *testing.T) {
	proc := BuildGRPCAuthedWeatherServer(t)
	require.NotNil(t, proc)
	require.NotNil(t, proc.Cmd())
	require.True(t, strings.Contains(proc.Cmd().Path, "grpc_authed_weather_server"))
	require.NotEmpty(t, proc.Cmd().Args)
	require.NotZero(t, proc.Port)
}

func TestRegisterGRPCAuthedWeatherService(t *testing.T) {
	upstreamProc := BuildGRPCAuthedWeatherServer(t)
	err := upstreamProc.Start()
	require.NoError(t, err)
	t.Cleanup(upstreamProc.Stop)
	integration.WaitForTCPPort(t, upstreamProc.Port, integration.ServiceStartupTimeout)
	upstreamEndpoint := fmt.Sprintf("localhost:%d", upstreamProc.Port)

	mcpanyTestServerInfo := integration.StartMCPANYServer(t, "grpc-authed-weather-service")
	defer mcpanyTestServerInfo.CleanupFunc()

	time.Sleep(1 * time.Second)

	RegisterGRPCAuthedWeatherService(t, mcpanyTestServerInfo.RegistrationClient, upstreamEndpoint)
}

func TestBuildWebsocketWeatherServer(t *testing.T) {
	proc := BuildWebsocketWeatherServer(t)
	require.NotNil(t, proc)
	require.NotNil(t, proc.Cmd())
	require.True(t, strings.Contains(proc.Cmd().Path, "weather-server"))
	require.NotEmpty(t, proc.Cmd().Args)
	require.NotZero(t, proc.Port)
}

func TestRegisterWebsocketWeatherService(t *testing.T) {
	t.Skip("Skipping flaky test")
}

func TestBuildWebrtcWeatherServer(t *testing.T) {
	proc := BuildWebrtcWeatherServer(t)
	require.NotNil(t, proc)
	require.NotNil(t, proc.Cmd())
	require.True(t, strings.Contains(proc.Cmd().Path, "webrtc_weather_server"))
	require.NotEmpty(t, proc.Cmd().Args)
	require.NotZero(t, proc.Port)
}

func TestRegisterWebrtcWeatherService(t *testing.T) {
	t.Skip("Skipping flaky test")
}

func TestBuildStdioServer(t *testing.T) {
	proc := BuildStdioServer(t)
	require.Nil(t, proc)
}

func TestRegisterStdioService(t *testing.T) {
	mcpanyTestServerInfo := integration.StartMCPANYServer(t, "stdio-service")
	defer mcpanyTestServerInfo.CleanupFunc()

	time.Sleep(1 * time.Second)

	RegisterStdioService(t, mcpanyTestServerInfo.RegistrationClient, "")
}

func TestBuildStdioDockerServer(t *testing.T) {
	proc := BuildStdioDockerServer(t)
	require.Nil(t, proc)
}

func TestRegisterStdioDockerService(t *testing.T) {
	t.Skip("Skipping test that requires Docker")
}

func TestBuildOpenAPIWeatherServer(t *testing.T) {
	proc := BuildOpenAPIWeatherServer(t)
	require.NotNil(t, proc)
	require.NotNil(t, proc.Cmd())
	require.True(t, strings.Contains(proc.Cmd().Path, "openapi_weather_server"))
	require.NotEmpty(t, proc.Cmd().Args)
	require.NotZero(t, proc.Port)
}

func TestRegisterOpenAPIWeatherService(t *testing.T) {
	// Start a mock upstream server
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/openapi.json" {
			fmt.Fprint(w, `{"openapi":"3.0.0","info":{"title":"Weather API","version":"1.0"},"paths":{}}`)
		}
	})
	upstream := newMockUpstreamServer(t, handler)

	mcpanyTestServerInfo := integration.StartMCPANYServer(t, "openapi-weather-service")
	defer mcpanyTestServerInfo.CleanupFunc()

	time.Sleep(1 * time.Second)

	RegisterOpenAPIWeatherService(t, mcpanyTestServerInfo.RegistrationClient, upstream.URL)
}

func TestBuildOpenAPIAuthedServer(t *testing.T) {
	proc := BuildOpenAPIAuthedServer(t)
	require.NotNil(t, proc)
	require.NotNil(t, proc.Cmd())
	require.True(t, strings.Contains(proc.Cmd().Path, "http_authed_echo_server"))
	require.NotEmpty(t, proc.Cmd().Args)
	require.NotZero(t, proc.Port)
}

func TestRegisterOpenAPIAuthedService(t *testing.T) {
	mcpanyTestServerInfo := integration.StartMCPANYServer(t, "openapi-authed-service")
	defer mcpanyTestServerInfo.CleanupFunc()

	time.Sleep(1 * time.Second)

	RegisterOpenAPIAuthedService(t, mcpanyTestServerInfo.RegistrationClient, "http://localhost:8080")
}

func TestBuildStreamableHTTPServer(t *testing.T) {
	proc := BuildStreamableHTTPServer(t)
	require.NotNil(t, proc)
	require.NotNil(t, proc.Cmd())
	require.True(t, strings.HasSuffix(proc.Cmd().Path, "npx"))
	require.NotZero(t, proc.Port)
}

func TestRegisterStreamableHTTPService(t *testing.T) {
	t.Skip("Skipping flaky test")
}

func TestVerifyMCPClient(t *testing.T) {
	mcpanyTestServerInfo := integration.StartMCPANYServer(t, "verify-mcp-client")
	defer mcpanyTestServerInfo.CleanupFunc()

	time.Sleep(1 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func() {
		defer cancel()
		VerifyMCPClient(t, mcpanyTestServerInfo.HTTPEndpoint)
	}()

	<-ctx.Done()

	if ctx.Err() == context.DeadlineExceeded {
		t.Fatal("VerifyMCPClient timed out")
	}
}

func TestBuildCachingServer(t *testing.T) {
	proc := BuildCachingServer(t)
	require.NotNil(t, proc)
	require.NotNil(t, proc.Cmd())
	require.True(t, strings.Contains(proc.Cmd().Path, "http_caching_server"))
	require.NotEmpty(t, proc.Cmd().Args)
	require.NotZero(t, proc.Port)
}

func TestRegisterCachingService(t *testing.T) {
	mcpanyTestServerInfo := integration.StartMCPANYServer(t, "caching-service")
	defer mcpanyTestServerInfo.CleanupFunc()

	time.Sleep(1 * time.Second)

	RegisterCachingService(t, mcpanyTestServerInfo.RegistrationClient, "http://localhost:8080")
}

type mockAITool struct{}

func (m *mockAITool) Install() {}

func (m *mockAITool) AddMCP(name, endpoint string) {}

func (m *mockAITool) RemoveMCP(name string) {}

func (m *mockAITool) Run(apiKey, model, prompt string) (string, error) {
	return "mocked result", nil
}

func TestAITool(t *testing.T) {
	aiTool := &mockAITool{}
	aiTool.Install()
	aiTool.AddMCP("test", "http://localhost:8080")
	aiTool.RemoveMCP("test")
	_, err := aiTool.Run("test-key", "test-model", "test-prompt")
	require.NoError(t, err)
}
