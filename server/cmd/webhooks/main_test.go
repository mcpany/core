// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	wd, _ := os.Getwd()
	fmt.Printf("Running tests in: %s\n", wd)

	// In Bazel, use the pre-built binary from runfiles.
	if srcDir := os.Getenv("TEST_SRCDIR"); srcDir != "" {
		workspace := os.Getenv("TEST_WORKSPACE")
		if workspace == "" {
			workspace = "_main"
		}
		bazelBin := filepath.Join(srcDir, workspace, "server", "cmd", "webhooks", "webhooks")
		if _, err := os.Stat(bazelBin); err == nil {
			// Symlink/copy to expected name
			_ = os.Symlink(bazelBin, "webhook-sidecar")
			code := m.Run()
			os.Remove("webhook-sidecar")
			os.Exit(code)
		}
	}

	// Fallback: build with go build (for non-Bazel runs)
	build := exec.Command("go", "build", "-o", "webhook-sidecar", ".")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build webhook-sidecar: %s\n%s", err, out)
		os.Exit(1)
	}

	code := m.Run()
	os.Remove("webhook-sidecar")
	os.Exit(code)
}

func TestWebhookSidecar(t *testing.T) {
	// Start the server on a random port
	port := "8092"
	cmd := exec.Command("./webhook-sidecar", "-port", port)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start webhook-sidecar: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
	}()

	// Wait for server to start
	url := fmt.Sprintf("http://localhost:%s", port)
	waitForServer(t, url+"/healthz")

	// Test /healthz
	resp, err := http.Get(url + "/healthz")
	if err != nil {
		t.Fatalf("failed to GET /healthz: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	// Test /markdown
	payload := `{"specversion":"1.0","type":"com.mcpany.webhook.request","source":"test","id":"123","data":{"result":"<h1>Hello</h1>"}}`
	req, _ := http.NewRequest("POST", url+"/markdown", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/cloudevents+json")

	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to POST /markdown: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp2.StatusCode)
	}
}

func waitForServer(t *testing.T, url string) {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("server failed to start within deadline")
}
