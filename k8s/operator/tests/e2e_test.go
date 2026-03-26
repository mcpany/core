// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package tests contains e2e tests for the MCP Operator.
package tests

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	clusterName = "mcp-e2e"
	kindImage   = "kindest/node:v1.27.3"
	namespace   = "mcp-system"
)

var tag = "1.0.0"

func init() {
	if t := os.Getenv("IMAGE_TAG"); t != "" {
		tag = t
	}
}

func TestOperatorE2E(t *testing.T) {
	if os.Getenv("E2E") != "true" {
		t.Skip("Skipping E2E test. Set E2E=true to run.")
	}

	checkPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	rootDir, err := getRootDir()
	if err != nil {
		t.Skipf("Skipping E2E test: %v", err)
	}
	t.Logf("Project root detected: %s", rootDir)

	if clusterExists(ctx, t, clusterName) {
		t.Logf("Deleting existing cluster %s to ensure clean state...", clusterName)
		runCommand(ctx, t, rootDir, "kind", "delete", "cluster", "--name", clusterName)
	}

	hostPort, err := getFreePort()
	if err != nil {
		t.Fatalf("Failed to get free port: %v", err)
	}
	t.Logf("Using host port %d for UI access (mapped to NodePort 30000)", hostPort)

	kindConfigContent := fmt.Sprintf(`kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  ipFamily: ipv4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 30000
    hostPort: %d
    listenAddress: "0.0.0.0"
    protocol: TCP
`, hostPort)
	tmpConfig := filepath.Join(t.TempDir(), "kind-config.yaml")
	if err := os.WriteFile(tmpConfig, []byte(kindConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write temp kind config: %v", err)
	}

	if err := runCommand(ctx, t, rootDir, "kind", "create", "cluster", "--name", clusterName, "--image", kindImage, "--config", tmpConfig, "--wait", "2m"); err != nil {
		t.Fatalf("Failed to create kind cluster: %v", err)
	}

	ensureBazelImageLoaded(t, filepath.Join("server", "cmd", "server", "server_tarball.sh"), "mcpany/server")
	ensureBazelImageLoaded(t, filepath.Join("ui", "ui_tarball.sh"), "mcpany/ui")
	ensureBazelImageLoaded(t, filepath.Join("server", "tests", "integration", "cmd", "mocks", "http_echo_server", "http_echo_server_tarball.sh"), "mcpany/http-echo-server")
	if os.Getenv("SKIP_IMAGE_BUILD") != "true" {
		t.Logf("Building Docker images with tag %s...", tag)
		if err := runCommand(ctx, t, rootDir, "docker", "build", "-t", fmt.Sprintf("mcpany/operator:%s", tag), "-f", "k8s/operator/Dockerfile", "."); err != nil {
			t.Fatalf("Failed to build operator image: %v", err)
		}
	}

	if err := runCommand(ctx, t, rootDir, "kind", "load", "docker-image", fmt.Sprintf("mcpany/server:%s", tag), "--name", clusterName); err != nil {
		t.Fatalf("Failed to load server image: %v", err)
	}
	if err := runCommand(ctx, t, rootDir, "kind", "load", "docker-image", fmt.Sprintf("mcpany/operator:%s", tag), "--name", clusterName); err != nil {
		t.Fatalf("Failed to load operator image: %v", err)
	}
	if err := runCommand(ctx, t, rootDir, "kind", "load", "docker-image", fmt.Sprintf("mcpany/ui:%s", tag), "--name", clusterName); err != nil {
		t.Fatalf("Failed to load ui image: %v", err)
	}
	if err := runCommand(ctx, t, rootDir, "kind", "load", "docker-image", "mcpany/http-echo-server:latest", "--name", clusterName); err != nil {
		t.Fatalf("Failed to load http-echo-server image: %v", err)
	}

	if err := runCommand(ctx, t, rootDir, "helm", "upgrade", "--install", "mcpany", "k8s/helm/mcpany",
		"--namespace", namespace,
		"--create-namespace",
		"--set", "image.repository=mcpany/server",
		"--set", fmt.Sprintf("image.tag=%s", tag),
		"--set", "image.pullPolicy=Never",
		"--set", "operator.enabled=true",
		"--set", "operator.image.repository=mcpany/operator",
		"--set", fmt.Sprintf("operator.image.tag=%s", tag),
		"--set", "operator.image.pullPolicy=Never",
		"--set", fmt.Sprintf("ui.image.tag=%s", tag),
		"--set", "ui.image.pullPolicy=Never",
		"--set", "ui.service.type=NodePort",
		"--set", "ui.service.nodePort=30000",
		"--set", "ui.apiKey=test-token",
		"--set", "apiKey=test-token",
		"--set", "env.MCPANY_ADMIN_INIT_PASSWORD=password",
		"--set", "env.MCPANY_DANGEROUS_ALLOW_LOCAL_IPS=true",
		"--set", "env.MCPANY_ALLOW_LOOPBACK_RESOURCES=true",
		"--set-file", "config=server/config.minimal.yaml",
		"--wait",
		"--timeout", "10m",
	); err != nil {
		t.Fatalf("Failed to install helm chart: %v", err)
	}

	if err := runCommand(ctx, t, rootDir, "kubectl", "wait", "--for=condition=ready", "pod", "-l", "app.kubernetes.io/name=mcpany", "-n", namespace, "--timeout=60s"); err != nil {
		t.Fatalf("Failed to wait for pods: %v", err)
	}

	if err := runCommand(ctx, t, rootDir, "kubectl", "run", "ui-http-echo-server", "--image=mcpany/http-echo-server:latest", "--image-pull-policy=Never", "--restart=Always", "-n", namespace); err != nil {
		t.Fatalf("Failed to deploy http-echo-server: %v", err)
	}
	if err := runCommand(ctx, t, rootDir, "kubectl", "expose", "pod", "ui-http-echo-server", "--port=5678", "--target-port=8080", "-n", namespace); err != nil {
		t.Fatalf("Failed to expose http-echo-server: %v", err)
	}
	if err := runCommand(ctx, t, rootDir, "kubectl", "wait", "--for=condition=ready", "pod", "ui-http-echo-server", "-n", namespace, "--timeout=60s"); err != nil {
		t.Fatalf("Failed to wait for http-echo-server: %v", err)
	}

	if err := waitForPort(ctx, t, fmt.Sprintf("127.0.0.1:%d", hostPort), 60*time.Second); err != nil {
		t.Fatalf("NodePort failed to become accessible: %v", err)
	}

	uiDir := filepath.Join(rootDir, "ui")
	workers := "4"
	if w := os.Getenv("PLAYWRIGHT_WORKERS"); w != "" {
		workers = w
	}
	playwrightArgs := []string{"test", "--workers=" + workers}
	if grep := os.Getenv("PLAYWRIGHT_GREP"); grep != "" {
		playwrightArgs = append(playwrightArgs, "--grep", grep)
	}
	if grepInvert := os.Getenv("PLAYWRIGHT_GREP_INVERT"); grepInvert != "" {
		playwrightArgs = append(playwrightArgs, "--grep-invert", grepInvert)
	}
	args := append([]string{"playwright"}, playwrightArgs...)
	playwrightCmd := exec.CommandContext(ctx, "npx", args...)
	playwrightCmd.Dir = uiDir
	playwrightCmd.Env = append(os.Environ(), fmt.Sprintf("PLAYWRIGHT_BASE_URL=http://127.0.0.1:%d", hostPort), "SKIP_WEBSERVER=true")
	playwrightCmd.Stdout = os.Stdout
	playwrightCmd.Stderr = os.Stderr

	if err := playwrightCmd.Run(); err != nil {
		t.Fatalf("UI Tests failed: %v", err)
	}
}

func checkPrerequisites(t *testing.T) {
	deps := []string{"kind", "kubectl", "helm", "docker"}
	for _, dep := range deps {
		if _, err := exec.LookPath(dep); err != nil {
			t.Skipf("Skipping E2E test: %s is not installed", dep)
		}
	}
}

func clusterExists(ctx context.Context, _ *testing.T, name string) bool {
	cmd := exec.CommandContext(ctx, "kind", "get", "clusters")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	clusters := strings.Split(string(out), "\n")
	for _, c := range clusters {
		if c == name {
			return true
		}
	}
	return false
}

func getRootDir() (string, error) {
	workspace := os.Getenv("TEST_WORKSPACE")
	if workspace == "" {
		workspace = "_main"
	}
	for _, candidate := range []string{
		os.Getenv("MCPANY_PROJECT_ROOT"),
		os.Getenv("GITHUB_WORKSPACE"),
		os.Getenv("BUILD_WORKSPACE_DIRECTORY"),
		filepath.Join(os.Getenv("TEST_SRCDIR"), workspace),
		filepath.Join(os.Getenv("RUNFILES_DIR"), "_main"),
	} {
		if candidate == "" {
			continue
		}
		if isProjectRoot(candidate) {
			return candidate, nil
		}
	}
	dir, _ := os.Getwd()
	for i := 0; i < 10; i++ {
		if isProjectRoot(dir) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("could not find project root")
}

func isProjectRoot(dir string) bool {
	if dir == "" {
		return false
	}
	if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(dir, "Makefile")); err == nil {
		if _, err := os.Stat(filepath.Join(dir, "server")); err == nil {
			return true
		}
	}
	return false
}

func ensureBazelImageLoaded(t *testing.T, loaderRelPath, imageName string) {
	t.Helper()
	workspace := os.Getenv("TEST_WORKSPACE")
	if workspace == "" {
		workspace = "_main"
	}
	for _, base := range []string{os.Getenv("TEST_SRCDIR"), os.Getenv("RUNFILES_DIR")} {
		if base == "" {
			continue
		}
		loader := filepath.Join(base, workspace, loaderRelPath)
		if _, err := os.Stat(loader); err != nil {
			continue
		}
		cmd := exec.Command(loader)
		cmd.Env = os.Environ()
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to load Bazel-built %s image: %v\n%s", imageName, err, string(out))
		}
		return
	}
}

func runCommand(ctx context.Context, _ *testing.T, dir string, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "DOCKER_API_VERSION=1.44")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func waitForPort(ctx context.Context, _ *testing.T, addr string, timeout time.Duration) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeoutTimer.C:
			return fmt.Errorf("timeout waiting for %s", addr)
		case <-ticker.C:
			conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
			if err == nil {
				conn.Close()
				return nil
			}
		}
	}
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", ":0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
