// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// Build the binary
	// We assume we are running from server/cmd/webhooks directory or with correct relative path
	// But `go test ./server/cmd/webhooks` runs inside a temp dir? No, usually in package dir.
	// Let's rely on `go build` with absolute path or relative to module root if possible.
	// Safer: Build explicitly targeting the package.

	// Check if we are in the right directory
	wd, _ := os.Getwd()
	fmt.Printf("Running tests in: %s\n", wd)

	build := exec.Command("go", "build", "-o", "webhook-sidecar", ".")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build webhook-sidecar: %s\n%s", err, out)
		os.Exit(1)
	}

	code := m.Run()

	// Clean up
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
