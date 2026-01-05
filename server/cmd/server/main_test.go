// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/appconsts"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// mockRunner is a mock implementation of the app.Runner interface for testing.
type mockRunner struct {
	called                   bool
	capturedStdio            bool
	capturedMcpListenAddress string
	capturedGrpcPort         string
	capturedConfigPaths      []string
	capturedShutdownTimeout  time.Duration
}

func (m *mockRunner) Run(_ context.Context, _ afero.Fs, stdio bool, mcpListenAddress, grpcPort string, configPaths []string, shutdownTimeout time.Duration) error {
	m.called = true
	m.capturedStdio = stdio
	m.capturedMcpListenAddress = mcpListenAddress
	m.capturedGrpcPort = grpcPort
	m.capturedConfigPaths = configPaths
	m.capturedShutdownTimeout = shutdownTimeout
	return nil
}

func (m *mockRunner) ReloadConfig(_ afero.Fs, _ []string) error {
	return nil
}

func (m *mockRunner) RunHealthServer(_ string) error {
	return nil
}

func TestHealthCmd(t *testing.T) {
	viper.Reset()
	// Start a mock HTTP server on a custom port
	portVal := findFreePort(t)
	port := fmt.Sprintf("%d", portVal)
	server := &http.Server{
		Addr:              ":" + port,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	}
	go func() {
		_ = server.ListenAndServe()
	}()
	defer func() { _ = server.Shutdown(context.Background()) }()

	// Wait for the server to start
	time.Sleep(100 * time.Millisecond)

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"health", "--mcp-listen-address", port})
	err := rootCmd.Execute()

	assert.NoError(t, err, "Health check should pass when server is running on custom port")
}

func TestHealthCmdWithCustomPort(t *testing.T) {
	viper.Reset()
	// Start a mock HTTP server on a custom port
	portVal := findFreePort(t)
	port := fmt.Sprintf("%d", portVal)
	server := &http.Server{
		Addr:              ":" + port,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	}
	go func() {
		_ = server.ListenAndServe()
	}()
	defer func() { _ = server.Shutdown(context.Background()) }()

	// Wait for the server to start
	time.Sleep(100 * time.Millisecond)

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"health", "--mcp-listen-address", port})
	err := rootCmd.Execute()

	assert.NoError(t, err, "Health check should pass when server is running on custom port")
}

func TestRootCmd(t *testing.T) {
	viper.Reset()
	mock := &mockRunner{}
	originalRunner := appRunner
	appRunner = mock
	defer func() { appRunner = originalRunner }()

	// Create temp config files/dirs that actually exist, otherwise fsnotify watcher fails
	tmpDir := t.TempDir()
	tmpFile, err := os.CreateTemp(tmpDir, "config*.yaml")
	assert.NoError(t, err)
	tmpFilePath := tmpFile.Name()
	_ = tmpFile.Close()

	port := findFreePort(t)
	grpcPort := findFreePort(t)

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{
		"run",
		"--stdio",
		"--mcp-listen-address", fmt.Sprintf("localhost:%d", port),
		"--grpc-port", fmt.Sprintf("%d", grpcPort),
		"--config-path", fmt.Sprintf("%s,%s", tmpFilePath, tmpDir),
		"--shutdown-timeout", "10s",
	})
	_ = rootCmd.Execute()

	assert.True(t, mock.called, "app.Run should have been called")
	assert.True(t, mock.capturedStdio, "stdio flag should be true")
	assert.Equal(t, fmt.Sprintf("localhost:%d", port), mock.capturedMcpListenAddress, "mcp-listen-address should be captured")
	assert.Equal(t, fmt.Sprintf("%d", grpcPort), mock.capturedGrpcPort, "grpc-port should be captured")
	assert.Equal(t, []string{tmpFilePath, tmpDir}, mock.capturedConfigPaths, "config-path should be captured")
	assert.Equal(t, 10*time.Second, mock.capturedShutdownTimeout, "shutdown-timeout should be captured")
}

func TestVersionCmd(t *testing.T) {
	viper.Reset()
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"version"})
	_ = rootCmd.Execute()

	_ = w.Close()
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
	viper.Reset()
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
	viper.Reset()
	// Start a mock HTTP server on a custom port
	portVal := findFreePort(t)
	port := fmt.Sprintf("%d", portVal)
	server := &http.Server{
		Addr:              ":" + port,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
	defer func() { _ = os.RemoveAll(dir) }()

	configFile := dir + "/config.yaml"
	err = os.WriteFile(configFile, []byte(`
global_settings:
  bind_address: "localhost:9090"
`), 0o600)
	assert.NoError(t, err)

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"health", "--config-path", configFile, "--mcp-listen-address", port})
	err = rootCmd.Execute()

	assert.NoError(t, err, "Health check should pass because the --mcp-listen-address flag should take precedence over the config file")
}

func findFreePort(t *testing.T) int {
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

func TestGracefulShutdown(t *testing.T) {
	if os.Getenv("GO_TEST_GRACEFUL_SHUTDOWN") == "1" {
		port := findFreePort(t)
		cmd := newRootCmd()
		cmd.SetArgs([]string{"run", "--mcp-listen-address", fmt.Sprintf("localhost:%d", port), "--config-path", ""})
		go func() {
			err := cmd.Execute()
			assert.NoError(t, err)
		}()
		// Wait for the server to start by polling the health check endpoint.
		assert.Eventually(t, func() bool {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/healthz", port))
			if err != nil {
				return false
			}
			defer func() { _ = resp.Body.Close() }()
			return resp.StatusCode == http.StatusOK
		}, 5*time.Second, 100*time.Millisecond)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestGracefulShutdown$") //nolint:gosec
	cmd.Env = append(os.Environ(), "GO_TEST_GRACEFUL_SHUTDOWN=1")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	err := cmd.Start()
	assert.NoError(t, err)

	// This is a bit of a hack, but we need to wait for the server to start
	// before we can get its port.
	time.Sleep(500 * time.Millisecond)

	// Send the interrupt signal to the child process group.
	err = syscall.Kill(-cmd.Process.Pid, syscall.SIGINT)
	assert.NoError(t, err)

	err = cmd.Wait()
	assert.NoError(t, err)
}

func TestConfigValidateCmd(t *testing.T) {
	viper.Reset()
	// Create a temporary valid config file
	validConfigFile, err := os.CreateTemp("", "valid-config-*.yaml")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(validConfigFile.Name()) }()
	_, err = validConfigFile.WriteString(`
global_settings:
  mcp_listen_address: "localhost:8080"
`)
	assert.NoError(t, err)
	_ = validConfigFile.Close()

	// Create a temporary invalid config file
	invalidConfigFile, err := os.CreateTemp("", "invalid-config-*.yaml")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(invalidConfigFile.Name()) }()
	_, err = invalidConfigFile.WriteString(`
invalid-yaml
`)
	assert.NoError(t, err)
	_ = invalidConfigFile.Close()

	// Create a temporary invalid config file (logic error)
	invalidLogicConfigFile, err := os.CreateTemp("", "invalid-logic-config-*.yaml")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(invalidLogicConfigFile.Name()) }()
	_, err = invalidLogicConfigFile.WriteString(`
upstream_services:
  - name: "my-service"
    mcp_service: {} # Missing connection type
`)
	assert.NoError(t, err)
	_ = invalidLogicConfigFile.Close()

	tests := []struct {
		name           string
		args           []string
		expectError    bool
		expectedOutput string
	}{
		{
			name:           "valid config",
			args:           []string{"config", "validate", "--config-path", validConfigFile.Name()},
			expectError:    false,
			expectedOutput: "Configuration is valid.\n",
		},
		{
			name:        "invalid config",
			args:        []string{"config", "validate", "--config-path", invalidConfigFile.Name()},
			expectError: true,
		},
		{
			name:        "invalid logic config",
			args:        []string{"config", "validate", "--config-path", invalidLogicConfigFile.Name()},
			expectError: true,
		},
		{
			name:           "no config file",
			args:           []string{"config", "validate"},
			expectError:    false,
			expectedOutput: "Configuration is valid.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			rootCmd := newRootCmd()
			rootCmd.SetArgs(tt.args)
			err := rootCmd.Execute()

			_ = w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = originalStdout

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, string(out))
			}
		})
	}
}

func TestConfigGenerateCmd(t *testing.T) {
	viper.Reset()
	originalStdout := os.Stdout
	originalStdin := os.Stdin
	defer func() {
		os.Stdout = originalStdout
		os.Stdin = originalStdin
	}()

	rOut, wOut, _ := os.Pipe()
	rIn, wIn, _ := os.Pipe()
	os.Stdout = wOut
	os.Stdin = rIn

	// Inputs for HTTP service generation
	inputs := []string{
		"http",
		"my-service",
		"http://localhost:8080",
		"get-data",
		"Get some data",
		"HTTP_METHOD_GET",
		"/data",
	}

	go func() {
		defer func() { _ = wIn.Close() }()
		for _, input := range inputs {
			_, _ = fmt.Fprintln(wIn, input)
		}
	}()

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"config", "generate"})
	err := rootCmd.Execute()

	_ = wOut.Close()
	out, _ := io.ReadAll(rOut)

	assert.NoError(t, err)
	output := string(out)
	assert.Contains(t, output, "MCP Any CLI: Configuration Generator")
	assert.Contains(t, output, "Generated configuration:")
	assert.Contains(t, output, "upstreamServices:")
	assert.Contains(t, output, "name: \"my-service\"")
}

func TestDocCmd(t *testing.T) {
	viper.Reset()
	// Create a temporary valid config file
	configFile, err := os.CreateTemp("", "doc-config-*.yaml")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(configFile.Name()) }()
	_, err = configFile.WriteString(`
upstream_services:
  - name: "my-service"
    http_service:
      address: "http://localhost:8080"
      tools:
        - name: "my-tool"
          call_id: "my-call"
      calls:
        my-call:
          id: "my-call"
          method: HTTP_METHOD_GET
          endpoint_path: "/test"
`)
	assert.NoError(t, err)
	_ = configFile.Close()

	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"config", "doc", "--config-path", configFile.Name()})
	err = rootCmd.Execute()

	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = originalStdout

	assert.NoError(t, err)
	output := string(out)
	assert.Contains(t, output, "# Available Tools")
	assert.Contains(t, output, "my-service.my-tool")
}

func TestUpdateCmd(t *testing.T) {
	viper.Reset()
	// Mock GitHub API
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/mcpany/core/releases/latest" {
			w.Header().Set("Content-Type", "application/json")
			// Return a release with same version as current
			fmt.Fprintf(w, `{"tag_name": "%s"}`, Version)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	t.Setenv("GITHUB_API_URL", ts.URL)

	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"update"})
	err := rootCmd.Execute()

	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = originalStdout

	assert.NoError(t, err)
	assert.Contains(t, string(out), "You are already running the latest version.")
}
