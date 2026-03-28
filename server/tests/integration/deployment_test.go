// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/stretchr/testify/require"
)

// dockerComposeDemoAPIKey is the API key defined in server/examples/docker-compose-demo/config.yaml.
const dockerComposeDemoAPIKey = "demo-key"

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func getDockerCommand(t *testing.T) []string {
	t.Helper()
	if os.Getenv("USE_SUDO_FOR_DOCKER") == "true" {
		return []string{"sudo", "docker"}
	}
	return []string{"docker"}
}

// helmChartDir returns the path to the k8s/helm/mcpany Helm chart directory.
// Under Bazel it resolves via runfiles; outside Bazel it falls back to the
// path relative to ProjectRoot.
func helmChartDir(t *testing.T) string {
	t.Helper()
	workspace := os.Getenv("TEST_WORKSPACE")
	if workspace == "" {
		workspace = "_main"
	}
	for _, base := range []string{os.Getenv("TEST_SRCDIR"), os.Getenv("RUNFILES_DIR")} {
		if base == "" {
			continue
		}
		candidate := filepath.Join(base, workspace, "k8s", "helm", "mcpany")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	// Fall back to path relative to project root
	rootDir := integration.ProjectRoot(t)
	return filepath.Join(rootDir, "../k8s", "helm", "mcpany")
}

// dockerComposeDir returns the path to the server/examples/docker-compose-demo directory.
// Under Bazel it resolves via runfiles; outside Bazel it falls back to the
// path relative to ProjectRoot.
func dockerComposeDir(t *testing.T) string {
	t.Helper()
	workspace := os.Getenv("TEST_WORKSPACE")
	if workspace == "" {
		workspace = "_main"
	}
	for _, base := range []string{os.Getenv("TEST_SRCDIR"), os.Getenv("RUNFILES_DIR")} {
		if base == "" {
			continue
		}
		candidate := filepath.Join(base, workspace, "server", "examples", "docker-compose-demo")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	// Fall back to path relative to project root
	rootDir := integration.ProjectRoot(t)
	return filepath.Join(rootDir, "examples", "docker-compose-demo")
}

func TestDockerCompose(t *testing.T) {
	if !integration.IsDockerSocketAccessible() {
	}
	if !commandExists("docker") {
	}

	srcComposeDir := dockerComposeDir(t)
	dockerComposeFile := filepath.Join(srcComposeDir, "docker-compose.yml")
	if _, err := os.Stat(dockerComposeFile); err != nil {
	}

	// Copy docker-compose files to a real temp directory so that Docker can bind-mount
	// them without issues from Bazel's runfile symlinks.
	composeDir := t.TempDir()
	for _, fname := range []string{"docker-compose.yml", "config.yaml"} {
		srcPath := filepath.Join(srcComposeDir, fname)
		dstPath := filepath.Join(composeDir, fname)
		srcData, err := os.ReadFile(srcPath) //nolint:gosec
		require.NoError(t, err, "reading %s", fname)
		require.NoError(t, os.WriteFile(dstPath, srcData, 0644), "writing %s", fname) //nolint:gosec
	}
	dockerComposeFile = filepath.Join(composeDir, "docker-compose.yml")

	// Load the Bazel-built server and echo-server images when running under Bazel.
	// Outside Bazel the images (mcpany/server:latest and mcpany/http-echo-server:latest)
	// must be pre-built.
	integration.EnsureServerImageLoaded(t)
	integration.EnsureHTTPEchoServerImageLoaded(t)

	dockerCmd := getDockerCommand(t)
	hostPort := findFreePort(t)
	composeEnv := append(os.Environ(), fmt.Sprintf("HOST_PORT=%d", hostPort))

	// Run in detached mode using the compose file's directory as the project directory
	// so that relative volume mounts (./config.yaml) resolve correctly.
	upCmdArgs := append(dockerCmd, "compose", "--project-directory", composeDir, "-f", dockerComposeFile, "up", "-d")
	upCmd := exec.Command(upCmdArgs[0], upCmdArgs[1:]...) //nolint:gosec
	upCmd.Env = composeEnv
	upOutput, err := upCmd.CombinedOutput()
	require.NoError(t, err, "docker compose up -d should not fail: %s", string(upOutput))

	// Cleanup function to bring down the services
	t.Cleanup(func() {
		// Capture logs before cleanup if test failed
		if t.Failed() {
			t.Log("Test failed, capturing docker logs...")
			logsCmd := exec.Command("docker", "compose", "--project-directory", composeDir, "-f", dockerComposeFile, "logs", "--no-color", "--tail=100")
			out, _ := logsCmd.CombinedOutput()
			t.Logf("Docker Logs:\n%s", string(out))
		}

		t.Log("Cleaning up docker compose services...")
		downCmdArgs := append(dockerCmd, "compose", "--project-directory", composeDir, "-f", dockerComposeFile, "down", "--volumes")
		downCmd := exec.Command(downCmdArgs[0], downCmdArgs[1:]...) //nolint:gosec
		downCmd.Env = composeEnv
		downOutput, err := downCmd.CombinedOutput()
		if err != nil {
			t.Logf("Failed to run 'docker compose down': %s\n%s", err, string(downOutput))
		} else {
			t.Log("Docker compose services cleaned up successfully.")
		}
	})

	// Wait for the services to be healthy
	require.Eventually(t, func() bool {
		psCmdArgs := append(dockerCmd, "compose", "--project-directory", composeDir, "-f", dockerComposeFile, "ps", "--format", "json")
		psCmd := exec.Command(psCmdArgs[0], psCmdArgs[1:]...) //nolint:gosec
		psCmd.Env = composeEnv
		psOutput, err := psCmd.CombinedOutput()
		if err != nil {
			t.Logf("docker compose ps failed: %v", err)
			return false
		}

		var services []map[string]interface{}
		// The output is a stream of JSON objects, so we need to handle that.
		decoder := json.NewDecoder(bytes.NewReader(psOutput))
		for decoder.More() {
			var service map[string]interface{}
			if err := decoder.Decode(&service); err != nil {
				t.Logf("Failed to decode docker compose ps output: %v", err)
				return false
			}
			services = append(services, service)
		}

		if len(services) < 2 {
			return false
		}

		mcpanyReady := false
		echoReady := false
		for _, s := range services {
			name, okName := s["Name"].(string)
			health, okHealth := s["Health"].(string)
			if !okName || !okHealth {
				continue
			}
			if strings.Contains(name, "mcpany-server") && health == "healthy" {
				mcpanyReady = true
			}
			if strings.Contains(name, "http-echo-server") && health == "healthy" {
				echoReady = true
			}
		}
		return mcpanyReady && echoReady
	}, 2*time.Minute, 5*time.Second, "Docker services did not become healthy in time")

	// Make a request to the echo tool via mcpany.
	// The streamable MCP endpoint requires an initialized session before tools/call.
	// We retry the full initialize -> initialized -> tools/call flow until the upstream service is ready.
	initializePayload := `{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"docker-compose-test","version":"1.0.0"}},"id":1}`
	initializedPayload := `{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}`
	payload := `{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "docker-http-echo.echo", "arguments": {"message": "Hello from Docker!"}}, "id": 2}`
	baseURL := fmt.Sprintf("http://127.0.0.1:%d/mcp?api_key=%s", hostPort, dockerComposeDemoAPIKey)
	require.Eventually(t, func() bool {
		client := &http.Client{Timeout: 10 * time.Second}

		initReq, err := http.NewRequest("POST", baseURL, bytes.NewBufferString(initializePayload))
		if err != nil {
			t.Logf("failed to create initialize request: %v", err)
			return false
		}
		initReq.Header.Set("Content-Type", "application/json")
		initReq.Header.Set("Accept", "application/json, text/event-stream")

		initResp, err := client.Do(initReq)
		if err != nil {
			t.Logf("initialize request failed: %v", err)
			return false
		}
		defer func() { _ = initResp.Body.Close() }()

		if initResp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(initResp.Body)
			t.Logf("unexpected initialize status code: %d body=%q", initResp.StatusCode, string(bodyBytes))
			return false
		}

		sessionID := initResp.Header.Get("Mcp-Session-Id")
		if sessionID == "" {
			t.Log("initialize response did not include MCP session id")
			return false
		}

		initializedReq, err := http.NewRequest("POST", baseURL, bytes.NewBufferString(initializedPayload))
		if err != nil {
			t.Logf("failed to create initialized request: %v", err)
			return false
		}
		initializedReq.Header.Set("Content-Type", "application/json")
		initializedReq.Header.Set("Accept", "application/json, text/event-stream")
		initializedReq.Header.Set("Mcp-Session-Id", sessionID)

		initializedResp, err := client.Do(initializedReq)
		if err != nil {
			t.Logf("initialized notification failed: %v", err)
			return false
		}
		defer func() { _ = initializedResp.Body.Close() }()

		if initializedResp.StatusCode != http.StatusOK && initializedResp.StatusCode != http.StatusAccepted {
			bodyBytes, _ := io.ReadAll(initializedResp.Body)
			t.Logf("unexpected initialized status code: %d body=%q", initializedResp.StatusCode, string(bodyBytes))
			return false
		}

		req, err := http.NewRequest("POST", baseURL, bytes.NewBufferString(payload))
		if err != nil {
			t.Logf("failed to create tool call request: %v", err)
			return false
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		req.Header.Set("Mcp-Session-Id", sessionID)

		resp, err := client.Do(req)
		if err != nil {
			t.Logf("tool call request failed: %v", err)
			return false
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Logf("unexpected tool-call status code: %d body=%q", resp.StatusCode, string(bodyBytes))
			return false
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Logf("failed to read response body: %v", err)
			return false
		}

		// Handle SSE format: extract the JSON payload from the "data:" field.
		jsonBytes := bodyBytes
		if resp.Header.Get("Content-Type") == "text/event-stream" || bytes.HasPrefix(bodyBytes, []byte("event: ")) {
			for _, line := range bytes.Split(bodyBytes, []byte("\n")) {
				if bytes.HasPrefix(line, []byte("data: ")) {
					jsonBytes = bytes.TrimPrefix(line, []byte("data: "))
					break
				}
			}
		}

		var result map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &result); err != nil {
			t.Logf("failed to decode response (raw body: %q): %v", string(bodyBytes), err)
			return false
		}
		return result["result"] != nil
	}, 30*time.Second, 2*time.Second, "Failed to get a successful tool-call response from mcpany")
}

func TestHelmChart(t *testing.T) {
	if !commandExists("helm") {
	}
	t.Parallel()

	helmPath := helmChartDir(t)
	if _, err := os.Stat(helmPath); err != nil {
	}

	// 1. Lint the chart
	lintCmd := exec.Command("helm", "lint", ".")
	lintCmd.Dir = helmPath
	lintOutput, err := lintCmd.CombinedOutput()
	require.NoError(t, err, "helm lint should not fail: %s", string(lintOutput))

	// 2. Template the chart to ensure it renders correctly
	templateCmd := exec.Command("helm", "template", "mcpany-release", ".")
	templateCmd.Dir = helmPath
	templateOutput, err := templateCmd.CombinedOutput()
	require.NoError(t, err, "helm template should not fail: %s", string(templateOutput))

	// 3. Check for expected resources in the output
	outputStr := string(templateOutput)
	require.Contains(t, outputStr, "kind: Service", "Rendered template should contain a Service")
	require.Contains(t, outputStr, "name: mcpany-release", "Rendered template should contain the release name")
	require.Contains(t, outputStr, "kind: Deployment", "Rendered template should contain a Deployment")
	require.Contains(t, outputStr, "app.kubernetes.io/name: mcpany", "Rendered template should contain the app name label")
}

func TestK8sFullStack(t *testing.T) {
	if !commandExists("helm") {
	}

	helmPath := helmChartDir(t)
	if _, err := os.Stat(helmPath); err != nil {
	}

	t.Parallel()

	// 1. Lint the chart
	t.Run("HelmLint", func(t *testing.T) {
		lintCmd := exec.Command("helm", "lint", ".")
		lintCmd.Dir = helmPath
		out, err := lintCmd.CombinedOutput()
		require.NoError(t, err, "helm lint should not fail: %s", string(out))
	})

	// 2. Client-side template rendering with default values (avoids cluster connection requirement)
	t.Run("HelmDryRunDefault", func(t *testing.T) {
		cmd := exec.Command("helm", "template", "mcpany-test", ".",
			"--set", "apiKey=test-key",
		)
		cmd.Dir = helmPath
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "helm template with default values should not fail: %s", string(out))
		outputStr := string(out)
		require.Contains(t, outputStr, "kind: Deployment", "Template output should contain a Deployment")
		require.Contains(t, outputStr, "kind: Service", "Template output should contain a Service")
	})

	// 3. Client-side template rendering with operator enabled
	t.Run("HelmDryRunWithOperator", func(t *testing.T) {
		cmd := exec.Command("helm", "template", "mcpany-operator-test", ".",
			"--set", "operator.enabled=true",
			"--set", "apiKey=test-key",
		)
		cmd.Dir = helmPath
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "helm template with operator enabled should not fail: %s", string(out))
		outputStr := string(out)
		require.Contains(t, outputStr, "kind: Deployment", "Template output with operator should contain a Deployment")
	})

	// 4. Template rendering – verify key resources and labels
	t.Run("HelmTemplateResources", func(t *testing.T) {
		cmd := exec.Command("helm", "template", "mcpany-release", ".",
			"--set", "apiKey=test-key",
		)
		cmd.Dir = helmPath
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "helm template should not fail: %s", string(out))
		outputStr := string(out)
		require.Contains(t, outputStr, "kind: Service", "Template should contain a Service")
		require.Contains(t, outputStr, "name: mcpany-release", "Template should reference the release name")
		require.Contains(t, outputStr, "kind: Deployment", "Template should contain a Deployment")
		require.Contains(t, outputStr, "app.kubernetes.io/name: mcpany", "Template should contain the app name label")
	})

	// 5. Template rendering – UI component enabled
	t.Run("HelmTemplateWithUI", func(t *testing.T) {
		cmd := exec.Command("helm", "template", "mcpany-release", ".",
			"--set", "ui.enabled=true",
			"--set", "apiKey=test-key",
		)
		cmd.Dir = helmPath
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "helm template with UI enabled should not fail: %s", string(out))
	})

	// 6. Template rendering – custom image tag propagated
	t.Run("HelmTemplateCustomImageTag", func(t *testing.T) {
		cmd := exec.Command("helm", "template", "mcpany-release", ".",
			"--set", "image.tag=v1.2.3",
			"--set", "apiKey=test-key",
		)
		cmd.Dir = helmPath
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "helm template with custom image tag should not fail: %s", string(out))
		require.Contains(t, string(out), "v1.2.3", "Custom image tag should appear in the rendered template")
	})
}

func findFreePort(t *testing.T) int {
	t.Helper()
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to resolve tcp addr: %v", err)
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		t.Fatalf("failed to listen on tcp addr: %v", err)
	}
	defer func() { _ = l.Close() }()
	return l.Addr().(*net.TCPAddr).Port
}
