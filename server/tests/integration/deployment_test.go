package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/tests/integration"
	"github.com/stretchr/testify/require"
)

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

func TestDockerCompose(t *testing.T) {
	t.SkipNow()
	if !integration.IsDockerSocketAccessible() {
		t.Skip("Docker socket not accessible, skipping TestDockerCompose.")
	}
	if !commandExists("docker") {
		t.Skip("docker command not found, skipping TestDockerCompose.")
	}

	t.Parallel()

	rootDir := integration.ProjectRoot(t)
	dockerComposeFile := filepath.Join(rootDir, "docker-compose.yml")

	dockerCmd := getDockerCommand(t)

	// Build the images first to avoid race conditions
	buildCmdArgs := dockerCmd
	buildCmdArgs = append(buildCmdArgs, "compose", "-f", dockerComposeFile, "build")
	buildCmd := exec.Command(buildCmdArgs[0], buildCmdArgs[1:]...) //nolint:gosec
	buildCmd.Dir = rootDir
	buildOutput, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "docker compose build should not fail: %s", string(buildOutput))

	// Run in detached mode
	upCmdArgs := dockerCmd
	upCmdArgs = append(upCmdArgs, "compose", "-f", dockerComposeFile, "up", "-d")
	upCmd := exec.Command(upCmdArgs[0], upCmdArgs[1:]...) //nolint:gosec
	upCmd.Dir = rootDir
	upOutput, err := upCmd.CombinedOutput()
	require.NoError(t, err, "docker compose up -d should not fail: %s", string(upOutput))

	// Cleanup function to bring down the services
	t.Cleanup(func() {
		t.Log("Cleaning up docker compose services...")
		downCmdArgs := dockerCmd
		downCmdArgs = append(downCmdArgs, "compose", "-f", dockerComposeFile, "down", "--volumes")
		downCmd := exec.Command(downCmdArgs[0], downCmdArgs[1:]...) //nolint:gosec
		downCmd.Dir = rootDir
		downOutput, err := downCmd.CombinedOutput()
		if err != nil {
			t.Logf("Failed to run 'docker compose down': %s\n%s", err, string(downOutput))
		} else {
			t.Log("Docker compose services cleaned up successfully.")
		}
	})

	// Wait for the services to be healthy
	require.Eventually(t, func() bool {
		psCmdArgs := dockerCmd
		psCmdArgs = append(psCmdArgs, "compose", "-f", dockerComposeFile, "ps", "--format", "json")
		psCmd := exec.Command(psCmdArgs[0], psCmdArgs[1:]...) //nolint:gosec
		psCmd.Dir = rootDir
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

	// Make a request to the echo tool via mcpany
	payload := `{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "docker-http-echo/-/echo", "arguments": {"message": "Hello from Docker!"}}, "id": 1}`
	req, err := http.NewRequest("POST", "http://localhost:50050", bytes.NewBufferString(payload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	require.Eventually(t, func() bool {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err = client.Do(req)
		if err != nil {
			t.Logf("curl request failed: %v", err)
			return false
		}
		defer func() { _ = resp.Body.Close() }()
		return resp.StatusCode == http.StatusOK
	}, 30*time.Second, 2*time.Second, "Failed to get a successful response from mcpany")

	defer func() { _ = resp.Body.Close() }()
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	// Check the response
	require.NotNil(t, result["result"])
	resultMap := result["result"].(map[string]interface{})
	contentMap := resultMap["content"].(map[string]interface{})
	dataMap := contentMap["data"].(map[string]interface{})
	require.Contains(t, dataMap["message"], "Hello from Docker!")
}

func TestHelmChart(t *testing.T) {
	t.SkipNow()
	if !commandExists("helm") {
		t.Skip("helm command not found, skipping TestHelmChart.")
	}
	t.Parallel()

	helmChartPath := filepath.Join(integration.ProjectRoot(t), "helm", "mcpany")

	// 1. Lint the chart
	lintCmd := exec.Command("helm", "lint", ".")
	lintCmd.Dir = helmChartPath
	lintOutput, err := lintCmd.CombinedOutput()
	require.NoError(t, err, "helm lint should not fail: %s", string(lintOutput))

	// 2. Template the chart to ensure it renders correctly
	templateCmd := exec.Command("helm", "template", "mcpany-release", ".")
	templateCmd.Dir = helmChartPath
	templateOutput, err := templateCmd.CombinedOutput()
	require.NoError(t, err, "helm template should not fail: %s", string(templateOutput))

	// 3. Check for expected resources in the output
	outputStr := string(templateOutput)
	require.Contains(t, outputStr, "kind: Service", "Rendered template should contain a Service")
	require.Contains(t, outputStr, "name: mcpany-release", "Rendered template should contain the release name")
	require.Contains(t, outputStr, "kind: Deployment", "Rendered template should contain a Deployment")
	require.Contains(t, outputStr, "app.kubernetes.io/name: mcpany", "Rendered template should contain the app name label")
}
