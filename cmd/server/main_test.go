// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/appconsts"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

// mockRunner is a mock implementation of the app.Runner interface for testing.
type mockRunner struct {
	called                  bool
	capturedStdio           bool
	capturedMcpListenAddress string
	capturedGrpcPort        string
	capturedConfigPaths     []string
	capturedShutdownTimeout time.Duration
}

func (m *mockRunner) Run(ctx context.Context, fs afero.Fs, stdio bool, mcpListenAddress, grpcPort string, configPaths []string, shutdownTimeout time.Duration) error {
	m.called = true
	m.capturedStdio = stdio
	m.capturedMcpListenAddress = mcpListenAddress
	m.capturedGrpcPort = grpcPort
	m.capturedConfigPaths = configPaths
	m.capturedShutdownTimeout = shutdownTimeout
	return nil
}

func (m *mockRunner) RunHealthServer(mcpListenAddress string) error {
	return nil
}

func TestHealthCmd(t *testing.T) {
	// This is a basic test to ensure the command runs without panicking.
	// A more thorough test would involve setting up a mock HTTP server.
	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"health", "--mcp-listen-address", "50051"})
	err := rootCmd.Execute()
	// We expect an error because no server is running
	assert.Error(t, err)
}

func TestHealthCmdWithCustomPort(t *testing.T) {
	// Start a mock HTTP server on a custom port
	port := "8088"
	server := &http.Server{
		Addr: ":" + port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	}
	go func() {
		server.ListenAndServe()
	}()
	defer server.Shutdown(context.Background())

	// Wait for the server to start
	time.Sleep(100 * time.Millisecond)

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"health", "--mcp-listen-address", port})
	err := rootCmd.Execute()

	assert.NoError(t, err, "Health check should pass when server is running on custom port")
}

func TestRootCmd(t *testing.T) {
	mock := &mockRunner{}
	originalRunner := appRunner
	appRunner = mock
	defer func() { appRunner = originalRunner }()

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{
		"--stdio",
		"--mcp-listen-address", "8081",
		"--grpc-port", "8082",
		"--config-path", "/etc/config.yaml,/etc/conf.d",
		"--shutdown-timeout", "10s",
	})
	rootCmd.Execute()

	assert.True(t, mock.called, "app.Run should have been called")
	assert.True(t, mock.capturedStdio, "stdio flag should be true")
	assert.Equal(t, "localhost:8081", mock.capturedMcpListenAddress, "mcp-listen-address should be captured")
	assert.Equal(t, "8082", mock.capturedGrpcPort, "grpc-port should be captured")
	assert.Equal(t, []string{"/etc/config.yaml", "/etc/conf.d"}, mock.capturedConfigPaths, "config-path should be captured")
	assert.Equal(t, 10*time.Second, mock.capturedShutdownTimeout, "shutdown-timeout should be captured")
}

func TestVersionCmd(t *testing.T) {
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"version"})
	rootCmd.Execute()

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = originalStdout

	expectedVersion := appconsts.Version
	if expectedVersion == "" {
		expectedVersion = "dev"
	}
	expectedOutput := appconsts.Name + " version " + expectedVersion + "\n"
	assert.Equal(t, expectedOutput, string(out))
}

// This test is for the main function, which is not easily testable.
// We can, however, test the command execution.
func TestMainExecution(t *testing.T) {
	// This is a bit of a meta-test. We're just making sure that calling main()
	// doesn't panic. We can't really inspect the output without more refactoring.
	// We will rely on the other tests to validate the behavior of the commands.
	assert.NotPanics(t, func() {
		// We can't actually run main because it will block.
		// Instead, we test the command directly.
		cmd := newRootCmd()
		cmd.SetArgs([]string{"--help"})
		err := cmd.Execute()
		assert.NoError(t, err)
	})
}

func TestHealthCmdFlagPrecedence(t *testing.T) {
	// Start a mock HTTP server on a custom port
	port := "8089"
	server := &http.Server{
		Addr: ":" + port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	}
	go func() {
		_ = server.ListenAndServe()
	}()
	defer func() { _ = server.Shutdown(context.Background()) }()

	// Wait for the server to start
	time.Sleep(100 * time.Millisecond)

	// Create a temporary config file with a different port
	dir, err := os.MkdirTemp("", "test-config")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	configFile := dir + "/config.yaml"
	err = os.WriteFile(configFile, []byte(`
global_settings:
  bind_address: "localhost:9090"
`), 0644)
	assert.NoError(t, err)

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"health", "--config-path", configFile, "--mcp-listen-address", port})
	err = rootCmd.Execute()

	assert.NoError(t, err, "Health check should pass because the --mcp-listen-address flag should take precedence over the config file")
}
