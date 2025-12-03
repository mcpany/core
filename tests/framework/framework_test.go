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
	"os"
	"strings"
	"testing"

	"github.com/mcpany/core/tests/integration"
	"github.com/stretchr/testify/require"
)

func TestBuildEverythingServer(t *testing.T) {
	proc := BuildEverythingServer(t)
	require.NotNil(t, proc)
	err := proc.Start()
	require.NoError(t, err)
	defer proc.Stop()
	integration.WaitForTCPPort(t, proc.Port, integration.ServiceStartupTimeout)
}

func TestRegisterEverythingService(t *testing.T) {
	upstreamProc := BuildEverythingServer(t)
	err := upstreamProc.Start()
	require.NoError(t, err)
	defer upstreamProc.Stop()
	integration.WaitForTCPPort(t, upstreamProc.Port, integration.ServiceStartupTimeout)

	testServerInfo := integration.StartMCPANYServer(t, "everything-test")
	defer testServerInfo.CleanupFunc()

	upstreamEndpoint := fmt.Sprintf("http://localhost:%d", upstreamProc.Port)
	RegisterEverythingService(t, testServerInfo.RegistrationClient, upstreamEndpoint)
}

func TestGeminiCLI(t *testing.T) {
	cli := NewGeminiCLI(t)
	cli.Install()

	// Test AddMCP
	endpoint := "http://localhost:12345"
	cli.AddMCP("test-mcp", endpoint)

	// Test Run
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping gemini run test")
	}

	output, err := cli.Run(apiKey, "what is the weather in painting?")
	require.NoError(t, err)
	require.True(t, strings.Contains(output, "Connection failed"), "expected connection error message")

	// Test RemoveMCP
	cli.RemoveMCP("test-mcp")
}

func TestBuildHTTPEchoServer(t *testing.T) {
	proc := BuildHTTPEchoServer(t)
	require.NotNil(t, proc)
	err := proc.Start()
	require.NoError(t, err)
	defer proc.Stop()
	integration.WaitForTCPPort(t, proc.Port, integration.ServiceStartupTimeout)
}

func TestRegisterHTTPEchoService(t *testing.T) {
	testServerInfo := integration.StartMCPANYServer(t, "http-echo-test")
	defer testServerInfo.CleanupFunc()

	upstreamEndpoint := "http://localhost:12345"
	RegisterHTTPEchoService(t, testServerInfo.RegistrationClient, upstreamEndpoint)
}

func TestBuildHTTPAuthedEchoServer(t *testing.T) {
	proc := BuildHTTPAuthedEchoServer(t)
	require.NotNil(t, proc)
	err := proc.Start()
	require.NoError(t, err)
	defer proc.Stop()
	integration.WaitForTCPPort(t, proc.Port, integration.ServiceStartupTimeout)
}

func TestRegisterHTTPAuthedEchoService(t *testing.T) {
	testServerInfo := integration.StartMCPANYServer(t, "http-authed-echo-test")
	defer testServerInfo.CleanupFunc()

	upstreamEndpoint := "http://localhost:12345"
	RegisterHTTPAuthedEchoService(t, testServerInfo.RegistrationClient, upstreamEndpoint)
}

func TestCaching(t *testing.T) {
	// Start the caching server
	cachingProc := BuildCachingServer(t)
	err := cachingProc.Start()
	require.NoError(t, err)
	defer cachingProc.Stop()
	integration.WaitForTCPPort(t, cachingProc.Port, integration.ServiceStartupTimeout)
	upstreamEndpoint := fmt.Sprintf("http://localhost:%d", cachingProc.Port)

	// Start the MCPAny server
	mcpanyTestServerInfo := integration.StartMCPANYServer(t, "caching-test")
	defer mcpanyTestServerInfo.CleanupFunc()

	// Register the caching service
	RegisterCachingService(t, mcpanyTestServerInfo.RegistrationClient, upstreamEndpoint)

	// Validate caching
	ValidateCaching(t, mcpanyTestServerInfo.HTTPEndpoint, upstreamEndpoint)
}

func TestGetUpstreamMetrics(t *testing.T) {
	// Start the caching server
	cachingProc := BuildCachingServer(t)
	err := cachingProc.Start()
	require.NoError(t, err)
	defer cachingProc.Stop()
	integration.WaitForTCPPort(t, cachingProc.Port, integration.ServiceStartupTimeout)
	upstreamEndpoint := fmt.Sprintf("http://localhost:%d", cachingProc.Port)

	// Get metrics
	metrics := getUpstreamMetrics(t, upstreamEndpoint)
	require.NotNil(t, metrics)
	require.Equal(t, int64(0), metrics["counter"])

	// Call the tool
	_, err = http.Get(fmt.Sprintf("%s/", upstreamEndpoint))
	require.NoError(t, err)

	// Get metrics again
	metrics = getUpstreamMetrics(t, upstreamEndpoint)
	require.NotNil(t, metrics)
	require.Equal(t, int64(1), metrics["counter"])
}
