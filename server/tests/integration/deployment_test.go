// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	// t.Skip("Skipping heavy integration test TestDockerCompose (flaky in CI/env due to header/port issues)")
	// // t.SkipNow()
	if !integration.IsDockerSocketAccessible() {
		// t.Skip("Docker socket not accessible, skipping TestDockerCompose.")
	}
	if !commandExists("docker") {
		// t.Skip("docker command not found, skipping TestDockerCompose.")
	}

	// t.Parallel() removed to avoid port conflicts with hardcoded 50050 in docker-compose.yml

	rootDir := integration.ProjectRoot(t)
	dockerComposeFile := filepath.Join(rootDir, "examples/docker-compose-demo/docker-compose.yml")

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
		// Capture logs before cleanup if test failed
		if t.Failed() {
			t.Log("Test failed, capturing docker logs...")
			logsCmd := exec.Command("docker", "compose", "-f", dockerComposeFile, "logs", "--no-color", "--tail=100")
			logsCmd.Dir = rootDir
			out, _ := logsCmd.CombinedOutput()
			t.Logf("Docker Logs:\n%s", string(out))
		}

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
	var resp *http.Response
	require.Eventually(t, func() bool {
		req, err := http.NewRequest("POST", "http://127.0.0.1:50050/mcp", bytes.NewBufferString(payload))
		if err != nil {
			t.Logf("failed to create request: %v", err)
			return false
		}
		req.Header.Set("Content-Type", "application/json")
		// Server requires Accept header to contain application/json (and potentially text/event-stream for SSE support check)
		req.Header.Set("Accept", "application/json, text/event-stream")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err = client.Do(req)
		if err != nil {
			t.Logf("curl request failed: %v", err)
			return false
		}
		if resp.StatusCode != http.StatusOK {
			t.Logf("Status code: %d", resp.StatusCode)
			_ = resp.Body.Close()
			return false
		}
		return true
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
	// // t.Skip("Skipping heavy integration test TestHelmChart")
	// Add build/env/bin to PATH to find helm installed by make
	rootDir := integration.ProjectRoot(t)
	buildBin := filepath.Join(rootDir, "../build/env/bin")
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", buildBin+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	if !commandExists("helm") {
		t.Skip("helm command not found, skipping TestHelmChart.")
	}
	t.Parallel()

	helmChartPath := filepath.Join(integration.ProjectRoot(t), "../k8s", "helm", "mcpany")

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

func TestK8sFullStack(t *testing.T) {
	if os.Getenv("E2E") != "true" {
		t.Skip("Skipping K8s E2E test (E2E=true not set)")
	}

	// Dependencies check
	requiredCmds := []string{"kind", "helm", "kubectl", "docker", "npm"}
	for _, cmd := range requiredCmds {
		if !commandExists(cmd) {
			t.Fatalf("%s command not found", cmd)
		}
	}

	rootDir := integration.ProjectRoot(t)
	helmChartPath := filepath.Join(rootDir, "../k8s", "helm", "mcpany")
	uiDir := filepath.Join(rootDir, "../ui")

	// Cluster Config
	clusterName := "mcpany-e2e-" + time.Now().Format("20060102-150405")
	t.Logf("Creating Kind cluster: %s", clusterName)

	// Create Cluster
	createClusterCmd := exec.Command("kind", "create", "cluster", "--name", clusterName)
	if out, err := createClusterCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to create kind cluster: %v\nOutput: %s", err, string(out))
	}

	// Cleanup
	t.Cleanup(func() {
		if t.Failed() {
			// Dump logs
			t.Log("Test failed, dumping cluster logs...")
			exec.Command("kubectl", "--context", "kind-"+clusterName, "get", "pods", "-o", "wide").Run()
			exec.Command("kubectl", "--context", "kind-"+clusterName, "describe", "pods").Run()
			exec.Command("kubectl", "--context", "kind-"+clusterName, "logs", "-l", "app.kubernetes.io/name=mcpany", "--tail=100").Run()
		}
		t.Logf("Deleting Kind cluster: %s", clusterName)
		exec.Command("kind", "delete", "cluster", "--name", clusterName).Run()
	})

	// Use the cluster context
	kubectlCtx := "kind-" + clusterName

	// Build Images
	t.Log("Building Docker images...")
	// Assuming make is available and works from root
	// We might need to run specific make targets or direct docker build commands
	// For simplicity calling 'make docker-build-all' from root might be heavy/slow
	// Let's build server and ui specifically or assume they are pre-built?
	// Better to build them here to ensure latest code is tested
	buildCmd := exec.Command("make", "-C", filepath.Join(rootDir, ".."), "docker-build-all")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build images: %v\nOutput: %s", err, string(out))
	}

	// Load Images into Kind
	// images := []string{"mcpany/server:latest", "mcpany/ui:latest", "redis:7", "postgres:15"}
	// We need to pull postgres/redis if not present locally, 'kind load' needs them locally
	// Or we just load mcpany images and let kind pull others?
	// Kind can pull public images. We only need to load local ones.
	localImages := []string{"mcpany/server:latest", "mcpany/ui:latest"}
	for _, img := range localImages {
		t.Logf("Loading image into Kind: %s", img)
		loadCmd := exec.Command("kind", "load", "docker-image", img, "--name", clusterName)
		if out, err := loadCmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to load image %s: %v\nOutput: %s", img, err, string(out))
		}
	}

	// Install Helm Chart
	t.Log("Installing Helm chart...")
	installCmd := exec.Command("helm", "install", "mcpany-e2e", helmChartPath,
		"--kube-context", kubectlCtx,
		"--set", "image.pullPolicy=Never", // For local kind images
		"--set", "image.tag=latest",
		"--set", "ui.image.pullPolicy=Never",
		"--set", "database.persistence.enabled=false", // Use ephemeral storage for test reliability
		"--set", "database.image.pullPolicy=IfNotPresent", // These come from registry
		"--set", "redis.image.pullPolicy=IfNotPresent",
		"--wait", // Wait for pods to be ready
		"--timeout", "5m")
	if out, err := installCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to install helm chart: %v\nOutput: %s", err, string(out))
	}

	// Port Forwarding
	// We need to forward UI port to access from Playwright
	// UI Service: mcpany-e2e-ui:3000
	uiSvcName := "mcpany-e2e-ui"

	// Start port-forward in background
	// We pick a random free port for UI
	// Doing simple port forward here
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	// uiPort := "3000" // Local port matches container port for simplicity, or find free one
	// uiPort := "3000" // Local port matches container port for simplicity, or find free one
	localUIPort := fmt.Sprintf("%d", findFreePort(t))

	pfCmd := exec.Command("kubectl", "--context", kubectlCtx, "port-forward", "svc/"+uiSvcName, localUIPort+":3000")
	if err := pfCmd.Start(); err != nil {
		t.Fatalf("Failed to start port-forward: %v", err)
	}
	defer func() {
		_ = pfCmd.Process.Kill()
	}()

	// Wait for port-forward compatibility
	t.Log("Waiting for port-forward to be ready...")
	require.Eventually(t, func() bool {
		resp, err := http.Get("http://127.0.0.1:" + localUIPort)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode < 500
	}, 60*time.Second, 1*time.Second, "UI did not become accessible via port-forward")

	// Run Playwright Tests
	t.Log("Running Playwright E2E tests...")
	e2eCmd := exec.Command("npm", "run", "test")
	e2eCmd.Dir = uiDir
	e2eCmd.Env = os.Environ()
	e2eCmd.Env = append(e2eCmd.Env, "REAL_CLUSTER=true")
	e2eCmd.Env = append(e2eCmd.Env, "PLAYWRIGHT_BASE_URL=http://127.0.0.1:"+localUIPort)
	// We might need to skip some tests specifically?
	// The current suite runs all tests in e2e.spec.ts.

	if out, err := e2eCmd.CombinedOutput(); err != nil {
		t.Fatalf("Playwright tests failed: %v\nOutput: %s", err, string(out))
	}

	t.Log("K8s Full Stack Test passed!")
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
