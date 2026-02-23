//go:build e2e

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
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

// TestProcessE2E replaces TestDockerComposeE2E with a local process-based test
// ensuring the stack logic is verified without Docker dependency.
func TestProcessE2E(t *testing.T) {
	binPath := BuildServer(t)
	rootDir := findRootDir(t)

	// 1. Start Echo Server (Mocking external service)
	echoURL := StartEchoServer(t)
	t.Logf("Echo server started at %s", echoURL)

	// 2. Config for Server
	configDir := filepath.Join(rootDir, "build", "e2e_process_stack")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)
	defer os.RemoveAll(configDir)

	configContent := fmt.Sprintf(`
upstream_services:
  - name: "docker-http-echo"
    http_service:
      address: "%s"
      calls:
        echo:
          method: "HTTP_METHOD_POST"
          endpoint_path: "/echo"
          parameters:
            - schema:
                name: "message"
                type: "STRING"
      tools:
        - name: "echo"
          description: "Echoes back the request body."
          call_id: "echo"

global_settings:
  # api_key omitted to use flag from StartServerProcess
`, echoURL)
	configPath := filepath.Join(configDir, "config.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// 3. Start Server
	t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", "true")
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")
	sp := StartServerProcess(t, binPath, "--config-path", configPath, "--debug")
	defer sp.Stop()

	// 4. Verify Health
	verifyEndpoint(t, fmt.Sprintf("%s/healthz", sp.BaseURL), 200, 30*time.Second)

	// 5. Functional Test: Simulate Gemini CLI & Verify Metrics
	t.Log("Simulating Gemini CLI interaction with echo tool...")
	simulateGeminiCLI(t, sp.BaseURL, sp.APIKey)

	t.Log("Verifying tool execution metrics...")
	verifyToolMetricDirect(t, fmt.Sprintf("%s/metrics", sp.BaseURL), "docker-http-echo.echo", sp.APIKey)

	// 6. Functional Test: Weather Service (Real external call)
	// We assume we can access internet. If not, this might fail, but let's try.
	t.Log("Starting Weather Service functional test...")
	testFunctionalWeatherLocal(t, rootDir, binPath)

	t.Log("E2E Test Passed!")
}

func testFunctionalWeatherLocal(t *testing.T, rootDir, binPath string) {
	// 1. Start mcpany-server with wttr.in config
	// marketplace is at repo root, rootDir is server root.
	configPath := filepath.Join(rootDir, "../marketplace/catalog/wttr.in/config.yaml")
	if _, err := os.Stat(configPath); err != nil {
		// Try rootDir if marketplace is inside server (backup)
		configPath = filepath.Join(rootDir, "marketplace/catalog/wttr.in/config.yaml")
		if _, err := os.Stat(configPath); err != nil {
			t.Fatalf("Config file not found at %s: %v", configPath, err)
		}
	}

	// Start Server
	t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", "true")
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")
	sp := StartServerProcess(t, binPath, "--config-path", configPath, "--api-key", "demo-key")
	defer sp.Stop()

	baseURL := sp.BaseURL

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
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-weather-client", Version: "1.0"}, nil)
	transport := &mcp.StreamableClientTransport{
		Endpoint: baseURL + "/mcp?api_key=demo-key",
	}

	session, err := client.Connect(ctx, transport, nil)
	require.NoError(t, err, "Failed to connect to MCP server for weather test")
	defer func() { _ = session.Close() }()

	// List tools to find the correct name
	var list *mcp.ListToolsResult
	var toolName string

	// Retry listing tools until get_weather appears or timeout
	// Service registration might be async or slow (DNS, etc)
	start := time.Now()
	for time.Since(start) < 30*time.Second {
		list, err = session.ListTools(ctx, nil)
		require.NoError(t, err, "Failed to list tools")

		var toolNames []string
		for _, tool := range list.Tools {
			toolNames = append(toolNames, tool.Name)
			if strings.Contains(tool.Name, "wttrin") && strings.HasSuffix(tool.Name, "get_weather") {
				toolName = tool.Name
			}
		}

		if toolName != "" {
			t.Logf("Found weather tool: %s", toolName)
			break
		}

		t.Logf("Waiting for weather tool... Available: %v", toolNames)
		time.Sleep(1 * time.Second)
	}

	if toolName == "" {
		t.Logf("Available tools (final): %v", list.Tools)
		require.Fail(t, "weather tool not found after waiting period")
	}
	t.Logf("Using tool name: %s", toolName)

	// Call get_weather with retry
	t.Log("Calling get_weather tool...")
	var result *mcp.CallToolResult

	// Retry up to 3 times
	for i := 0; i < 3; i++ {
		result, err = session.CallTool(ctx, &mcp.CallToolParams{
			Name:      toolName,
			Arguments: json.RawMessage(`{"location": "London"}`),
		})
		if err == nil {
			break
		}
		t.Logf("CallTool failed (attempt %d/3): %v. Retrying...", i+1, err)
		time.Sleep(1 * time.Second)
	}
	require.NoError(t, err, "Failed to call get_weather tool after retries")

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
		req, err := http.NewRequest("GET", metricsURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("X-API-Key", "demo-key")

		resp, err := http.DefaultClient.Do(req)
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
	cmd.Env = os.Environ() // Explicitly pass environment to ensure t.Setenv works
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	t.Logf("Running: %s %s (Env: COMPOSE_PROJECT_NAME=%s)", name, strings.Join(args, " "), os.Getenv("COMPOSE_PROJECT_NAME"))
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
func simulateGeminiCLI(t *testing.T, baseURL, apiKey string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	transport := &mcp.StreamableClientTransport{
		Endpoint: baseURL + "/mcp?api_key=" + apiKey,
	}

	session, err := client.Connect(ctx, transport, nil)
	require.NoError(t, err, "Failed to connect to MCP server")
	defer func() { _ = session.Close() }()

	// 1. Initialize is handled by Connect

	// List tools to find the correct name
	list, err := session.ListTools(ctx, nil)
	require.NoError(t, err, "Failed to list tools")

	var toolNames []string
	for _, tool := range list.Tools {
		toolNames = append(toolNames, tool.Name)
	}
	t.Logf("Available tools: %v", toolNames)

	toolName := "echo"
	for _, tool := range list.Tools {
		if strings.HasSuffix(tool.Name, "echo") {
			toolName = tool.Name
			break
		}
	}
	t.Logf("Using tool name: %s", toolName)

	// 2. Send generic tool call with retry
	t.Log("Calling echo tool...")
	var result *mcp.CallToolResult
	for i := 0; i < 3; i++ {
		result, err = session.CallTool(ctx, &mcp.CallToolParams{
			Name:      toolName,
			Arguments: json.RawMessage(`{"message": "Hello MCP"}`),
		})
		if err == nil {
			break
		}
		t.Logf("CallTool failed (attempt %d/3): %v. Retrying...", i+1, err)
		time.Sleep(2 * time.Second)
	}
	require.NoError(t, err, "Failed to call echo tool after retries")

	// Check result
	require.NotNil(t, result)
	require.NotEmpty(t, result.Content)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected TextContent")
	require.Contains(t, textContent.Text, "Hello MCP")
}

// verifyToolMetricDirect verifies metrics directly from the text-based /metrics endpoint
// used when Prometheus is not available in the stack.
func verifyToolMetricDirect(t *testing.T, metricsURL, toolName, apiKey string) {
	// Retry for up to 5 seconds
	deadline := time.Now().Add(5 * time.Second)
	var body string

	for time.Now().Before(deadline) {
		req, err := http.NewRequest("GET", metricsURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("X-API-Key", apiKey)

		resp, err := http.DefaultClient.Do(req)
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

// createDynamicCompose creates a temporary docker-compose file with 0 ports in the build directory
func createDynamicCompose(t *testing.T, rootDir, originalPath string) string {
	content, err := os.ReadFile(originalPath)
	require.NoError(t, err)

	// Replace fixed ports with 0 ports
	// Match pattern: "HOST_PORT:CONTAINER_PORT"
	// We want to replace any HostPort with a specific port from our range.
	re := regexp.MustCompile(`"([0-9\$\{}:_a-zA-Z-]+):([0-9]+)"`)
	port := 25200
	s := re.ReplaceAllStringFunc(string(content), func(match string) string {
		parts := re.FindStringSubmatch(match)
		res := fmt.Sprintf(`"%d:%s"`, port, parts[2])
		port++
		return res
	})

	// Inject SSRF allow-lists into mcpany-server environment (first environment block)
	s = strings.Replace(s, "environment:", "environment:\n      - MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES=true\n      - MCPANY_DANGEROUS_ALLOW_LOCAL_IPS=true", 1)

	// Inject MCPANY_ENABLE_FILE_CONFIG=true into services
	if !strings.Contains(s, "MCPANY_ENABLE_FILE_CONFIG") {
		s = strings.Replace(s, "environment:", "environment:\n      - MCPANY_ENABLE_FILE_CONFIG=true", -1)
	}

	// Ensure build directory exists
	buildDir := filepath.Join(rootDir, "build")
	err = os.MkdirAll(buildDir, 0755)
	require.NoError(t, err)

	// Create temp file in build dir
	tmpFile, err := os.CreateTemp(buildDir, "docker-compose-dynamic-*.yml")
	require.NoError(t, err)

	_, err = tmpFile.WriteString(s)
	require.NoError(t, err)
	tmpFile.Close()

	return tmpFile.Name()
}
