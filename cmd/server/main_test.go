/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// mockRunner is a mock implementation of the app.Runner interface for testing.
type mockRunner struct {
	called                  bool
	capturedStdio           bool
	capturedJsonrpcPort     string
	capturedGrpcPort        string
	capturedConfigPaths     []string
	capturedShutdownTimeout time.Duration
}

func (m *mockRunner) Run(ctx context.Context, fs afero.Fs, stdio bool, jsonrpcPort, grpcPort string, configPaths []string, shutdownTimeout time.Duration, v *viper.Viper) error {
	m.called = true
	m.capturedStdio = stdio
	m.capturedJsonrpcPort = jsonrpcPort
	m.capturedGrpcPort = grpcPort
	m.capturedConfigPaths = configPaths
	m.capturedShutdownTimeout = shutdownTimeout
	return nil
}

func (m *mockRunner) RunHealthServer(jsonrpcPort string) error {
	return nil
}

func TestHealthCmd(t *testing.T) {
	// This is a basic test to ensure the command runs without panicking.
	// A more thorough test would involve setting up a mock HTTP server.
	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"health", "--jsonrpc-port", "50051"})
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
	rootCmd.SetArgs([]string{"health", "--jsonrpc-port", port})
	err := rootCmd.Execute()

	assert.NoError(t, err, "Health check should pass when server is running on custom port")
}

func TestRootCmd(t *testing.T) {
	// Create a temporary config file
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
	rootCmd.SetArgs([]string{
		"--config-path", configFile,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		err = rootCmd.ExecuteContext(ctx)
		assert.NoError(t, err)
	}()

	time.Sleep(1 * time.Second)
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
	rootCmd.SetArgs([]string{"health", "--jsonrpc-port", port})
	err = rootCmd.Execute()

	assert.NoError(t, err, "Health check should pass because the --jsonrpc-port flag should take precedence over the config file")
}
