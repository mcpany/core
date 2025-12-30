// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func findFreePortRepro(t *testing.T) int {
	t.Helper()
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
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

func TestStartupWithBrokenUpstream(t *testing.T) {
	if os.Getenv("GO_TEST_BROKEN_UPSTREAM") == "1" {
		port := findFreePortRepro(t)
		grpcPort := findFreePortRepro(t)

		// Create a config with a broken upstream (gRPC with reflection)
		configFile, err := os.CreateTemp("", "broken-config-*.yaml")
		assert.NoError(t, err)
		defer os.Remove(configFile.Name())

		configContent := `
upstream_services:
  - name: "broken-grpc"
    grpc_service:
      address: "localhost:54321"
      use_reflection: true
`
		_, err = configFile.WriteString(configContent)
		assert.NoError(t, err)
		configFile.Close()

		cmd := newRootCmd()
		cmd.SetArgs([]string{
			"run",
			"--mcp-listen-address", fmt.Sprintf("localhost:%d", port),
			"--grpc-port", fmt.Sprintf("localhost:%d", grpcPort),
			"--config-path", configFile.Name(),
		})

		// Run in a goroutine
		go func() {
			err := cmd.Execute()
			if err != nil {
				fmt.Printf("Server exited with error: %v\n", err)
			}
		}()

		// Wait for the server to start by polling the health check endpoint.
		// If it crashes on startup due to broken upstream, this will timeout.
		assert.Eventually(t, func() bool {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/healthz", port))
			if err != nil {
				return false
			}
			defer func() { _ = resp.Body.Close() }()
			return resp.StatusCode == http.StatusOK
		}, 10*time.Second, 100*time.Millisecond)

		// Keep alive for logs
		time.Sleep(15 * time.Second)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestStartupWithBrokenUpstream$")
	cmd.Env = append(os.Environ(), "GO_TEST_BROKEN_UPSTREAM=1")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Capture stdout/stderr to check for logs
	var outBuf, errBuf bytes.Buffer
	// Use a thread-safe buffer or wrapper if needed, but here simple buffer is fine as we read after wait.
	// Actually we want to read continuously or poll?
	// assert.Eventually reads from the buffer, so we need concurrent access safety?
	// bytes.Buffer is NOT thread safe.
	// But exec.Command writes to it.
	// We can't use assert.Eventually on the buffer content easily while it's being written to without a mutex.
	// However, we can just wait for the process to finish (which it does after 15s sleep) and THEN check logs.

	cmd.Stdout = io.MultiWriter(os.Stdout, &outBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &errBuf)

	err := cmd.Start()
	assert.NoError(t, err)

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			// It might fail if the server exits with error, but we expect it to run 15s then exit?
			// The child process just sleeps and returns, so it should exit with 0.
			// Unless the server crashes.
			t.Logf("Test process exited: %v", err)
		}
	case <-time.After(20 * time.Second):
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		t.Fatal("Test timed out")
	}

	// Verify that we logged the error for the broken service
	logs := outBuf.String() + errBuf.String()
	assert.Contains(t, logs, "Failed to register service", "Logs should contain error about failed registration")
	assert.Contains(t, logs, "broken-grpc", "Logs should contain the name of the broken service")

	// Ensure we also see "HTTP server listening" to confirm it didn't crash
	assert.Contains(t, logs, "HTTP server listening", "Server should have started")
}
