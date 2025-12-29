// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package e2e_sequential

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestDockerComposeE2E(t *testing.T) {
	if os.Getenv("E2E_DOCKER") != "true" {
		t.Skip("Skipping E2E Docker test. Set E2E_DOCKER=true to run.")
	}

	rootDir, err := os.Getwd()
	require.NoError(t, err)

	// Navigate up if we are running from inside tests/e2e or tests/e2e_sequential
	if strings.HasSuffix(rootDir, "tests/e2e") || strings.HasSuffix(rootDir, "tests/e2e_sequential") {
		rootDir = filepath.Join(rootDir, "../..")
	}
	rootDir, err = filepath.Abs(rootDir)
	require.NoError(t, err)

	imageName := "ghcr.io/mcpany/server:latest"

	// 1. Build Docker Image
	t.Log("Building mcpany/server image...")
	runCommand(t, rootDir, "docker", "build", "-t", imageName, "-f", "docker/Dockerfile.server", ".")

	// Cleanup function
	cleanup := func() {
		t.Log("Cleaning up...")
		dumpLogs := func(name string) {
			cmd := exec.Command("docker", "logs", name)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			_ = cmd.Run()
		}
		dumpLogs("core-mcpany-server-1")
		dumpLogs("docker-compose-demo-mcpany-server-1")
		dumpLogs("mcpany-weather-test")

		_ = exec.Command("docker", "compose", "down", "-v").Run()
		_ = exec.Command("docker", "compose", "-f", "examples/docker-compose-demo/docker-compose.yml", "down", "-v").Run()
		_ = exec.Command("docker", "rm", "-f", "mcpany-weather-test").Run()
		time.Sleep(5 * time.Second)
	}
	defer cleanup()
	cleanup() // Ensure clean start

	// 2. Start Root Docker Compose (Production)
	if _, err := os.Stat(fmt.Sprintf("%s/docker-compose.yml", rootDir)); err == nil {
		t.Log("Starting root docker-compose...")
		runCommand(t, rootDir, "docker", "compose", "up", "-d", "--wait")

		// 3. Verify Health
		t.Log("Verifying mcpany-server health...")
		verifyEndpoint(t, "http://localhost:50050/healthz", 200, 30*time.Second)

		// 4. Verify Prometheus Metrics
		t.Log("Verifying Prometheus metrics...")
		// Wait a bit for scraping
		time.Sleep(15 * time.Second)
		verifyPrometheusMetric(t, "http://localhost:9099/api/v1/query?query=up", "mcpany-server:50050")
	} else {
		t.Log("Skipping root docker-compose test (docker-compose.yml not found)")
	}

	// 5. Start Example Docker Compose
	t.Log("Switching to example docker-compose...")
	// Stop root to avoid conflicts
	runCommand(t, rootDir, "docker", "compose", "down")

	exampleDir := "examples/docker-compose-demo"
	runCommand(t, rootDir, "docker", "compose", "-f", fmt.Sprintf("%s/docker-compose.yml", exampleDir), "up", "-d", "--wait")

	// 6. Verify Example Health
	t.Log("Verifying example mcpany-server health...")
	// Example server is also exposed on host 50050
	verifyEndpoint(t, "http://localhost:50050/healthz", 200, 30*time.Second)

	// 7. Functional Test: Simulate Gemini CLI & Verify Metrics
	t.Log("Simulating Gemini CLI interaction with echo tool...")
	simulateGeminiCLI(t, "http://localhost:50050")

	t.Log("Verifying tool execution metrics...")
	verifyToolMetricDirect(t, "http://localhost:50050/metrics", "docker-http-echo.echo")

	// 8. Functional Test: Weather Service (Real external call)
	t.Log("Starting Weather Service functional test...")
	testFunctionalWeather(t, rootDir)

	t.Log("E2E Test Passed!")
}

func testFunctionalWeather(t *testing.T, rootDir string) {
	// 1. Start mcpany-server with wttr.in config
	// We run it on a different port to avoid conflict with previous steps if they didn't clean up fully,
	// or just to be isolated.
	port := 50060
	// Use local config file instead of remote URL to ensure reliability
	configPath := fmt.Sprintf("%s/examples/popular_services/wttr.in/config.yaml", rootDir)
    t.Logf("rootDir: %s", rootDir)
    t.Logf("configPath: %s", configPath)
    if _, err := os.Stat(configPath); err != nil {
        t.Fatalf("Config file not found at %s: %v", configPath, err)
    }

	t.Logf("Starting mcpany-server for weather test on port %d...", port)

    // Read and log config file content to be sure
    content, rErr := os.ReadFile(configPath)
    if rErr == nil {
        t.Logf("Config file content:\n%s", string(content))
    } else {
        t.Logf("Failed to read config file: %v", rErr)
    }

	cmd := exec.Command("docker", "run", "-d", "--name", "mcpany-weather-test",
		"-p", fmt.Sprintf("%d:50050", port),
		"-v", fmt.Sprintf("%s:/config.yaml", configPath),
		"ghcr.io/mcpany/server:latest",
		"run", "--config-path", "/config.yaml", "--mcp-listen-address", ":50050",
	)
    t.Logf("Running command: %s", cmd.String())

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	require.NoError(t, err, "Failed to start weather server container")

	// Cleanup
	defer func() {
		_ = exec.Command("docker", "kill", "mcpany-weather-test").Run()
	}()

	baseURL := fmt.Sprintf("http://localhost:%d", port)

	// 2. Wait for health
	verifyEndpoint(t, fmt.Sprintf("%s/healthz", baseURL), 200, 30*time.Second)

	// 3. Simulate Gemini CLI interaction
	t.Log("Simulating Gemini CLI interaction with get_weather tool...")
	toolName := simulateGeminiCLIWeather(t, baseURL)

	// 4. Verify Metrics
	t.Log("Verifying tool execution metrics for weather service...")
	// Parse service_id from toolName (format: serviceID.toolName)
	parts := strings.Split(toolName, ".")
	if len(parts) < 2 {
		t.Fatalf("Unexpected tool name format: %s", toolName)
	}
	// Service ID is the prefix (everything before the last dot)
	serviceID := strings.Join(parts[:len(parts)-1], ".")

	t.Logf("Verifying metrics for tool: %s, service: %s", toolName, serviceID)
	verifyToolMetricWithService(t, fmt.Sprintf("%s/metrics", baseURL), toolName, serviceID)
}

func simulateGeminiCLIWeather(t *testing.T, baseURL string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-weather-client", Version: "1.0"}, nil)
	transport := &mcp.StreamableClientTransport{
		Endpoint: baseURL,
	}

	session, err := client.Connect(ctx, transport, nil)
	require.NoError(t, err, "Failed to connect to MCP server for weather test")
	defer func() { _ = session.Close() }()

	// List tools to find the correct name
	list, err := session.ListTools(ctx, nil)
	require.NoError(t, err, "Failed to list tools")
	t.Logf("Available tools: %v", list.Tools)

	toolName := "get_weather"
	for _, tool := range list.Tools {
		if strings.HasSuffix(tool.Name, "get_weather") {
			toolName = tool.Name
			break
		}
	}
	t.Logf("Using tool name: %s", toolName)

	// Call get_weather
	t.Log("Calling get_weather tool...")
	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: json.RawMessage(`{"location": "London"}`),
	})
	require.NoError(t, err, "Failed to call get_weather tool")

	require.NotNil(t, result)
	require.NotEmpty(t, result.Content)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected TextContent")
	t.Logf("Weather tool response: %s", textContent.Text)
	require.Contains(t, textContent.Text, "London", "Response should contain the location name")

	return toolName
}

func verifyToolMetricWithService(t *testing.T, metricsURL, toolName, serviceID string) {
	// Retry for up to 5 seconds
	deadline := time.Now().Add(5 * time.Second)
	var body string

	for time.Now().Before(deadline) {
		//nolint:gosec // G107: Url is constructed internally in test
		resp, err := http.Get(metricsURL)
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		body = string(bodyBytes)

		if strings.Contains(body, "mcpany_tools_call_total") &&
			strings.Contains(body, fmt.Sprintf("tool=\"%s\"", toolName)) &&
			strings.Contains(body, fmt.Sprintf("service_id=\"%s\"", serviceID)) {
			return // Success
		}

		time.Sleep(500 * time.Millisecond)
	}

	// Failed after retries, fail with detailed message
	t.Logf("Metrics output:\n%s", body)
	require.Contains(t, body, "mcpany_tools_call_total", "Metric name not found")
	require.Contains(t, body, fmt.Sprintf("tool=\"%s\"", toolName), "Tool label not found")
	require.Contains(t, body, fmt.Sprintf("service_id=\"%s\"", serviceID), "Service ID label not found")
}

func runCommand(t *testing.T, dir string, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	t.Logf("Running: %s %s", name, strings.Join(args, " "))
	err := cmd.Run()
	require.NoError(t, err, "Command failed: %s %s", name, strings.Join(args, " "))
}

func verifyEndpoint(t *testing.T, url string, expectedStatus int, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		//nolint:gosec // G107: Url is constructed internally in test
		resp, err := http.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == expectedStatus {
				return
			}
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatalf("Failed to verify endpoint %s within %v", url, timeout)
}

func verifyPrometheusMetric(t *testing.T, url string, expectedTarget string) {
	//nolint:gosec // G107: Url is constructed internally in test
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Define a struct to match Prometheus API response
	type PrometheusResponse struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Value  []interface{}     `json:"value"` // [timestamp, value]
			} `json:"result"`
		} `json:"data"`
	}

	var parsed PrometheusResponse
	err = json.Unmarshal(body, &parsed)
	require.NoError(t, err, "Failed to parse Prometheus response: %s", string(body))

	require.Equal(t, "success", parsed.Status)
	require.NotEmpty(t, parsed.Data.Result, "No metrics found in response")

	found := false
	for _, result := range parsed.Data.Result {
		val := result.Value[1].(string)
		if val == "1" {
			t.Logf("Found metric: %v value: %s", result.Metric, val)
			found = true
		}
	}
	require.True(t, found, "No 'up' metric with value '1' found")

	// Verify 'up' metric contains target
	require.Contains(t, string(body), expectedTarget)
}

// simulateGeminiCLI simulates a basic Gemini CLI interaction (MCP client)
// It connects via SSE and sends a JSON-RPC tool call using the Go SDK.
func simulateGeminiCLI(t *testing.T, baseURL string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	transport := &mcp.StreamableClientTransport{
		Endpoint: baseURL,
	}

	session, err := client.Connect(ctx, transport, nil)
	require.NoError(t, err, "Failed to connect to MCP server")
	defer func() { _ = session.Close() }()

	// 1. Initialize is handled by Connect

	// List tools to find the correct name
	list, err := session.ListTools(ctx, nil)
	require.NoError(t, err, "Failed to list tools")
	t.Logf("Available tools: %v", list.Tools)

	toolName := "echo"
	for _, tool := range list.Tools {
		if strings.HasSuffix(tool.Name, "echo") {
			toolName = tool.Name
			break
		}
	}
	t.Logf("Using tool name: %s", toolName)

	// 2. Send generic tool call
	t.Log("Calling echo tool...")
	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: json.RawMessage(`{"message": "Hello MCP"}`),
	})
	require.NoError(t, err, "Failed to call echo tool")

	// Check result
	require.NotNil(t, result)
	require.NotEmpty(t, result.Content)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected TextContent")
	require.Contains(t, textContent.Text, "Hello MCP")
}

// verifyToolMetricDirect verifies metrics directly from the text-based /metrics endpoint
// used when Prometheus is not available in the stack.
func verifyToolMetricDirect(t *testing.T, metricsURL, toolName string) {
	// Retry for up to 5 seconds
	deadline := time.Now().Add(5 * time.Second)
	var body string

	for time.Now().Before(deadline) {
		//nolint:gosec // G107: Url is constructed internally in test
		resp, err := http.Get(metricsURL)
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		body = string(bodyBytes)

		if strings.Contains(body, "mcpany_tools_call_total") &&
			strings.Contains(body, fmt.Sprintf("tool=\"%s\"", toolName)) {
			return // Success
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Failed
	t.Logf("Metrics output:\n%s", body)
	require.Contains(t, body, "mcpany_tools_call_total", "Metric name not found")
	require.Contains(t, body, fmt.Sprintf("tool=\"%s\"", toolName), "Tool label not found")
}
