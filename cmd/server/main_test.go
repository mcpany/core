// Copyright 2025 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0


package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"syscall"
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

func (m *mockRunner) Run(ctx context.Context, fs afero.Fs, stdio, watch bool, mcpListenAddress, grpcPort string, configPaths []string, shutdownTimeout time.Duration) error {
	m.called = true
	m.capturedStdio = stdio
	m.capturedMcpListenAddress = mcpListenAddress
	m.capturedGrpcPort = grpcPort
	m.capturedConfigPaths = configPaths
	m.capturedShutdownTimeout = shutdownTimeout
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

func TestValidateCommand(t *testing.T) {
	// Create a temporary directory for config files
	dir, err := os.MkdirTemp("", "test-validate")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	// Create a dummy valid configuration file
	validConfig := `
upstream_services:
  - name: "my-http-service"
    http_service:
      address: "https://api.example.com"
      tools:
        - name: "get_user"
          description: "Get user by ID"
          call_id: "get_user_call"
      calls:
        get_user_call:
          method: "HTTP_METHOD_GET"
          endpoint_path: "/users/{userId}"
`
	validConfigFile := dir + "/valid_config.yaml"
	err = os.WriteFile(validConfigFile, []byte(validConfig), 0644)
	assert.NoError(t, err)

	// Create a dummy invalid configuration file
	invalidConfig := `
upstream_services:
  - name: "my-http-service"
    http_service:
      address: "https://api.example.com"
      tools:
        - name: "get_user"
          description: "Get user by ID"
          call_id: "get_user_call"
      calls:
        get_user_call:
          method: "INVALID_METHOD"
          endpoint_path: "/users/{userId}"
`
	invalidConfigFile := dir + "/invalid_config.yaml"
	err = os.WriteFile(invalidConfigFile, []byte(invalidConfig), 0644)
	assert.NoError(t, err)

	testCases := []struct {
		name          string
		args          []string
		expectSuccess bool
		expectedError string
	}{
		{
			name:          "ValidConfig",
			args:          []string{"validate", "--config-path", validConfigFile},
			expectSuccess: true,
		},
		{
			name:          "InvalidConfig",
			args:          []string{"validate", "--config-path", invalidConfigFile},
			expectSuccess: false,
			expectedError: "invalid value for enum field method: \"INVALID_METHOD\"",
		},
		{
			name:          "NoConfigPath",
			args:          []string{"validate"},
			expectSuccess: false,
			expectedError: "no configuration paths provided",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rootCmd := newRootCmd()
			rootCmd.SetArgs(tc.args)
			err := rootCmd.Execute()

			if tc.expectSuccess {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			}
		})
	}
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
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func TestGracefulShutdown(t *testing.T) {
	if os.Getenv("GO_TEST_GRACEFUL_SHUTDOWN") == "1" {
		port := findFreePort(t)
		cmd := newRootCmd()
		cmd.SetArgs([]string{"--mcp-listen-address", fmt.Sprintf("localhost:%d", port)})
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
			defer resp.Body.Close()
			return resp.StatusCode == http.StatusOK
		}, 5*time.Second, 100*time.Millisecond)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestGracefulShutdown$")
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
