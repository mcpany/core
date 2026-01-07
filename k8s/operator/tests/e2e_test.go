// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	clusterName = "mcp-e2e"
	kindImage   = "kindest/node:v1.29.2"
	namespace   = "mcp-system"
	tag         = "1.0.0"
)

func TestOperatorE2E(t *testing.T) {
	if os.Getenv("E2E") != "true" {
		t.Skip("Skipping E2E test. Set E2E=true to run.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	rootDir, err := getRootDir()
	if err != nil {
		t.Fatalf("Failed to get root dir: %v", err)
	}
	t.Logf("Project root detected: %s", rootDir)

	// 1. Cleanup previous runs
	t.Log("Cleaning up previous runs...")
	runCommand(t, ctx, rootDir, "helm", "uninstall", "mcpany", "-n", namespace, "--wait")
	runCommand(t, ctx, rootDir, "kubectl", "delete", "ns", namespace, "--wait")

	// 2. Check prerequisites
	checkPrerequisites(t)

	// 3. Create Kind Cluster
	if !clusterExists(t, ctx, clusterName) {
		t.Logf("Creating Kind cluster %s...", clusterName)
		if err := runCommand(t, ctx, rootDir, "kind", "create", "cluster", "--name", clusterName, "--image", kindImage, "--config", "k8s/tests/kind-config.yaml", "--wait", "2m"); err != nil {
			t.Fatalf("Failed to create kind cluster: %v", err)
		}
	} else {
		t.Logf("Cluster %s already exists.", clusterName)
	}

	// 4. Build Images (Locally)
	t.Logf("Building Docker images with tag %s...", tag)
	// Prepare docker context for server build
	if err := runCommand(t, ctx, rootDir, "make", "-C", "server", "prepare-docker-context"); err != nil {
		t.Fatalf("Failed to prepare docker context: %v", err)
	}

	// Build server with root context because Dockerfile expects it (COPY server/go.mod etc.)
	if err := runCommand(t, ctx, rootDir, "make", "-C", "server", "build-docker", fmt.Sprintf("SERVER_IMAGE_TAGS=mcpany/server:%s", tag)); err != nil {
		t.Fatalf("Failed to build server image: %v", err)
	}
	if err := runCommand(t, ctx, rootDir, "docker", "build", "-t", fmt.Sprintf("mcpany/operator:%s", tag), "-f", "k8s/operator/Dockerfile", "."); err != nil {
		t.Fatalf("Failed to build operator image: %v", err)
	}
	if err := runCommand(t, ctx, rootDir, "docker", "build", "-t", fmt.Sprintf("mcpany/ui:%s", tag), "-f", "ui/Dockerfile", "."); err != nil {
		t.Fatalf("Failed to build ui image: %v", err)
	}

	// 5. Load Images into Kind
	t.Log("Loading images into Kind...")
	if err := runCommand(t, ctx, rootDir, "kind", "load", "docker-image", fmt.Sprintf("mcpany/server:%s", tag), "--name", clusterName); err != nil {
		t.Fatalf("Failed to load server image: %v", err)
	}
	if err := runCommand(t, ctx, rootDir, "kind", "load", "docker-image", fmt.Sprintf("mcpany/operator:%s", tag), "--name", clusterName); err != nil {
		t.Fatalf("Failed to load operator image: %v", err)
	}
	if err := runCommand(t, ctx, rootDir, "kind", "load", "docker-image", fmt.Sprintf("mcpany/ui:%s", tag), "--name", clusterName); err != nil {
		t.Fatalf("Failed to load ui image: %v", err)
	}

	// 6. Install Helm Chart
	t.Log("Installing Helm chart...")
	// Create namespace
	cmd := exec.CommandContext(ctx, "kubectl", "create", "namespace", namespace, "--dry-run=client", "-o", "yaml")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to create namespace yaml: %v", err)
	}
	applyCmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", "-")
	applyCmd.Stdin = strings.NewReader(string(out))
	if out, err := applyCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to apply namespace: %v, output: %s", err, string(out))
	}

	// Helm upgrade --install
	if err := runCommand(t, ctx, rootDir, "helm", "upgrade", "--install", "mcpany", "k8s/helm/mcpany",
		"--namespace", namespace,
		"--set", fmt.Sprintf("image.repository=mcpany/server"),
		"--set", fmt.Sprintf("image.tag=%s", tag),
		"--set", "image.pullPolicy=Never",
		"--set", "operator.enabled=true",
		"--set", "operator.image.repository=mcpany/operator",
		"--set", fmt.Sprintf("operator.image.tag=%s", tag),
		"--set", "operator.image.pullPolicy=Never",
		"--set", fmt.Sprintf("ui.image.tag=%s", tag),
		"--set", "ui.image.pullPolicy=Never",
		"--wait",
		"--timeout", "10m",
	); err != nil {
		t.Fatalf("Failed to install helm chart: %v", err)
	}

	t.Log("Deployment successful!")

	// 7. Verify Pods
	t.Log("Verifying pods...")
	if err := runCommand(t, ctx, rootDir, "kubectl", "wait", "--for=condition=ready", "pod", "-l", "app.kubernetes.io/name=mcpany", "-n", namespace, "--timeout=60s"); err != nil {
		t.Fatalf("Failed to wait for pods: %v", err)
	}
}

func checkPrerequisites(t *testing.T) {
	deps := []string{"kind", "kubectl", "helm", "docker"}
	for _, dep := range deps {
		if _, err := exec.LookPath(dep); err != nil {
			t.Fatalf("Error: %s is not installed", dep)
		}
	}
}

func clusterExists(t *testing.T, ctx context.Context, name string) bool {
	cmd := exec.CommandContext(ctx, "kind", "get", "clusters")
	out, err := cmd.Output()
	if err != nil {
		t.Logf("Failed to get clusters: %v", err)
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
	// Assuming test is run from k8s/operator/tests, go up 3 levels to find root
	// Or better, find go.mod file
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// Walk up until we find go.work, which should be in the root
	for i := 0; i < 10; i++ {
		// Check for go.work or Makefile which should be in root
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir, nil
		}
		if _, err := os.Stat(filepath.Join(dir, "Makefile")); err == nil {
			// Double check it has server directory to be sure
			if _, err := os.Stat(filepath.Join(dir, "server")); err == nil {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("could not find project root (go.work or Makefile+server) from %s", dir)
}

func runCommand(t *testing.T, ctx context.Context, dir string, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	t.Logf("Running: %s %v", name, args)
	return cmd.Run()
}
