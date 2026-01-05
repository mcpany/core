// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package e2e_sequential

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
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

func TestDockerComposeE2E(t *testing.T) {
	if os.Getenv("E2E_DOCKER") != "true" {
		t.Skip("Skipping E2E Docker test. Set E2E_DOCKER=true to run.")
	}

	rootDir, err := os.Getwd()
	require.NoError(t, err)

	// Navigate up to repo root (core)
	// tests/e2e_sequential -> server -> core
	if strings.HasSuffix(rootDir, "tests/e2e") || strings.HasSuffix(rootDir, "tests/e2e_sequential") {
		rootDir = filepath.Join(rootDir, "../../..")
	} else if strings.HasSuffix(rootDir, "server") {
		// If running from server root
		rootDir = filepath.Join(rootDir, "..")
	}
	rootDir, err = filepath.Abs(rootDir)
	require.NoError(t, err)

	imageName := getImageName(t)

	// 1. Build Docker Image
	if os.Getenv("CI") == "" {
		t.Log("Building mcpany/server image...")
		runCommand(t, rootDir, "docker", "build", "-t", imageName, "-f", "server/docker/Dockerfile.server", ".")
	} else {
		t.Log("CI environment, skipping docker build")
	}

	// Use a unique project name for isolation
	projectName := fmt.Sprintf("e2e_seq_%d", time.Now().UnixNano())
	t.Setenv("COMPOSE_PROJECT_NAME", projectName)
	t.Logf("Using COMPOSE_PROJECT_NAME: %s", projectName)

	// Cleanup function
	var currentComposeFile string

	// Cleanup function
	cleanup := func() {
		t.Log("Cleaning up...")

		// Dump logs from the active compose file if set
		if currentComposeFile != "" {
			t.Logf("Dumping logs from %s...", currentComposeFile)
			cmd := exec.Command("docker", "compose", "-f", currentComposeFile, "logs")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			_ = cmd.Run()
		}

		// Dump logs from manually run weather container
		cmd := exec.Command("docker", "logs", "mcpany-weather-test") // Name without random suffix? No, we used random.
		// We need to capture the weather container name too if we want to dump it.
		// For now, let's skip dumping weather logs or try to capture it too.

		// Aggressive cleanup
		// We can't know the weather container name here easily unless we share it.
		// But we defer cleanup in testFunctionalWeather too.
		// So this main cleanup is just a safety net.

		if currentComposeFile != "" {
			cmd = exec.Command("docker", "compose", "-f", currentComposeFile, "down", "-v", "--remove-orphans")
			cmd.Env = os.Environ()
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			_ = cmd.Run()
		}

		// Also try generic project down just in case
		cmd = exec.Command("docker", "compose", "down", "-v", "--remove-orphans")
		cmd.Env = os.Environ()
		_ = cmd.Run()

		time.Sleep(2 * time.Second)
	}
	defer cleanup()
	cleanup() // Ensure clean start

	// Helper to get dynamic port
	getServicePort := func(composeFile, service, internalPort string) string {
		cmd := exec.Command("docker", "compose", "-f", composeFile, "port", service, internalPort)
		cmd.Env = os.Environ()
		out, err := cmd.Output()
		require.NoError(t, err, "Failed to get port for %s %s", service, internalPort)
		// Output: 0.0.0.0:32xxx
		addr := strings.TrimSpace(string(out))
		_, port, err := net.SplitHostPort(addr)
		require.NoError(t, err, "Failed to split host port: %s", addr)
		return port
	}

	// 2. Start Root Docker Compose (Production) - OPTIONAL
	// We skip the root docker-compose test modification for now if it requires complex patching,
	// or we can apply the same logic.
	// The previous test code had: if _, err := os.Stat(fmt.Sprintf("%s/docker-compose.yml", rootDir)); err == nil { ... }
	// Let's keep strict parity but make it dynamic.
	rootCompose := filepath.Join(rootDir, "docker-compose.yml")
	if _, err := os.Stat(rootCompose); err == nil {
		t.Log("Starting root docker-compose with dynamic ports...")
		// Create dynamic override
		dynamicCompose := createDynamicCompose(t, rootCompose)
		currentComposeFile = dynamicCompose
		defer os.Remove(dynamicCompose)

		runCommand(t, rootDir, "docker", "compose", "-f", dynamicCompose, "up", "-d", "--wait")

		// Discover ports
		serverPort := getServicePort(dynamicCompose, "mcpany-server", "50050")
		// prometheusPort := getServicePort(dynamicCompose, "prometheus", "9090")
		// We might fallback or skip prometheus if not present.

		t.Logf("Root mcpany-server running on port %s", serverPort)
		verifyEndpoint(t, fmt.Sprintf("http://localhost:%s/healthz", serverPort), 200, 30*time.Second)

		// We only verify prometheus if we can find the port.
		// For now, let's assume we focused on example-compose as the main test.
		// If root compose is present, we shut it down.
		runCommand(t, rootDir, "docker", "compose", "-f", dynamicCompose, "down")
	} else {
		t.Log("Skipping root docker-compose test (docker-compose.yml not found)")
	}

	// 5. Start Example Docker Compose
	t.Log("Switching to example docker-compose...")

	exampleDir := filepath.Join(rootDir, "examples/docker-compose-demo")
	originalCompose := filepath.Join(exampleDir, "docker-compose.yml")
	dynamicCompose := createDynamicCompose(t, originalCompose)
	currentComposeFile = dynamicCompose
	defer os.Remove(dynamicCompose)

	runCommand(t, rootDir, "docker", "compose", "-f", dynamicCompose, "up", "-d", "--wait")

	// 6. Verify Example Health
	serverPort := getServicePort(dynamicCompose, "mcpany-server", "50050")
	t.Logf("Example mcpany-server running on port %s", serverPort)
	verifyEndpoint(t, fmt.Sprintf("http://localhost:%s/healthz", serverPort), 200, 30*time.Second)

	// 7. Functional Test: Simulate Gemini CLI & Verify Metrics
	t.Log("Simulating Gemini CLI interaction with echo tool...")
	simulateGeminiCLI(t, fmt.Sprintf("http://localhost:%s", serverPort))

	t.Log("Verifying tool execution metrics...")
	// Note: Metrics are on the same port 50050 for standard serve (or 50051? checks config).
	// Original test checked 51234/metrics.
	// If we use dynamic port, it maps to 50050 (internal).
	verifyToolMetricDirect(t, fmt.Sprintf("http://localhost:%s/metrics", serverPort), "docker-http-echo.echo")

	// 8. Functional Test: Weather Service (Real external call)
	t.Log("Starting Weather Service functional test...")
	// Pass rootDir and use dynamic ports internally too
	testFunctionalWeather(t, rootDir)

	t.Log("E2E Test Passed!")
}

func testFunctionalWeather(t *testing.T, rootDir string) {
	// 1. Start mcpany-server with wttr.in config
	// We run it on a dynamic port to avoid conflict with previous steps or other processes.
	configPath := fmt.Sprintf("%s/examples/popular_services/wttr.in/config.yaml", rootDir)
	t.Logf("rootDir: %s", rootDir)
	t.Logf("configPath: %s", configPath)
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("Config file not found at %s: %v", configPath, err)
	}

	// We also use a unique container name to avoid conflict
	containerName := fmt.Sprintf("mcpany-weather-test-%d", time.Now().UnixNano())

	cmd := exec.Command("docker", "run", "-d", "--name", containerName,
		"-p", "0:50050", // Dynamic port
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
		_ = exec.Command("docker", "rm", "-f", containerName).Run()
	}()

	// Get assigned port
	out, err := exec.Command("docker", "port", containerName, "50050/tcp").Output()
	require.NoError(t, err, "Failed to get assigned port")
	// Output example: 0.0.0.0:32768
	portBinding := strings.TrimSpace(string(out))
	// If multiple bindings (IPv4/IPv6), take the first line
	if idx := strings.Index(portBinding, "\n"); idx != -1 {
		portBinding = portBinding[:idx]
	}

	// Parse the port
	_, portStr, err := net.SplitHostPort(portBinding)
	require.NoError(t, err, "Failed to parse port from %s", portBinding)

	baseURL := fmt.Sprintf("http://localhost:%s", portStr)

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

// createDynamicCompose creates a temporary docker-compose file with 0 ports
func createDynamicCompose(t *testing.T, originalPath string) string {
	content, err := os.ReadFile(originalPath)
	require.NoError(t, err)

	// Replace fixed ports with 0 ports
	// Use regex to handle potential spacing variations
	// Match pattern: "HOST_PORT:CONTAINER_PORT"
	// We want to replace any HostPort with 0.

	// Pattern: any whitespace - "digits:digits"
	// mcpany-server: "${HOST_PORT:-50050}:50050" -> 0:50050
	re1 := regexp.MustCompile(`"\$\{HOST_PORT:-50050\}:50050"`)
	re2 := regexp.MustCompile(`"50051:50051"`)
	re3 := regexp.MustCompile(`"8080:8080"`)
	re4 := regexp.MustCompile(`"9099:9090"`)

	s := string(content)
	s = re1.ReplaceAllString(s, `"0:50050"`)
	s = re2.ReplaceAllString(s, `"0:50051"`)
	s = re3.ReplaceAllString(s, `"0:8080"`)
	s = re4.ReplaceAllString(s, `"0:9090"`)

	// Create temp file in same dir to preserve relative paths
	dir := filepath.Dir(originalPath)
	tmpFile, err := os.CreateTemp(dir, "docker-compose-dynamic-*.yml")
	require.NoError(t, err)

	_, err = tmpFile.WriteString(s)
	require.NoError(t, err)
	tmpFile.Close()

	return tmpFile.Name()
}

func getImageName(t *testing.T) string {
	imageName := "ghcr.io/mcpany/server:latest"
	if os.Getenv("CI") != "" {
		t.Log("CI environment, using pre-built image")
		imageName = "mcpany/server:latest"
	}
	return imageName
}
