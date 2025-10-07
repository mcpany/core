/*
 * Copyright 2025 Author(s) of MCP-XY
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
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/mcpxy/core/pkg/appconsts"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

// mockRunner is a mock implementation of the app.Runner interface for testing.
type mockRunner struct {
	called              bool
	capturedStdio       bool
	capturedJsonrpcPort string
	capturedGrpcPort    string
	capturedConfigPaths []string
	returnErr           error
}

func (m *mockRunner) Run(ctx context.Context, fs afero.Fs, stdio bool, jsonrpcPort, grpcPort string, configPaths []string) error {
	m.called = true
	m.capturedStdio = stdio
	m.capturedJsonrpcPort = jsonrpcPort
	m.capturedGrpcPort = grpcPort
	m.capturedConfigPaths = configPaths
	return m.returnErr
}

func TestRootCmd(t *testing.T) {
	mock := &mockRunner{}
	originalRunner := appRunner
	appRunner = mock
	defer func() { appRunner = originalRunner }()

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{
		"--stdio",
		"--jsonrpc-port", "8081",
		"--grpc-port", "8082",
		"--config-paths", "/etc/config.yaml,/etc/conf.d",
	})
	rootCmd.Execute()

	assert.True(t, mock.called, "app.Run should have been called")
	assert.True(t, mock.capturedStdio, "stdio flag should be true")
	assert.Equal(t, "8081", mock.capturedJsonrpcPort, "jsonrpc-port should be captured")
	assert.Equal(t, "8082", mock.capturedGrpcPort, "grpc-port should be captured")
	assert.Equal(t, []string{"/etc/config.yaml", "/etc/conf.d"}, mock.capturedConfigPaths, "config-paths should be captured")
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

func TestMainExitCodeOnFailure(t *testing.T) {
	// This sub-process pattern is a standard way to test `main` functions in Go.
	// The test is re-run with an environment variable set to trigger the `main` function.
	if os.Getenv("GO_TEST_SUBPROCESS") == "1" {
		appRunner = &mockRunner{returnErr: fmt.Errorf("simulated run error")}
		main()
		return
	}

	// The main test process executes the test binary as a subprocess.
	cmd := exec.Command(os.Args[0], "-test.run=^TestMainExitCodeOnFailure$")
	cmd.Env = append(os.Environ(), "GO_TEST_SUBPROCESS=1")

	// We expect the buggy version to exit with code 0, so `cmd.Run()` will return nil.
	// After the fix, it should exit with a non-zero code, and `cmd.Run()` will return an error.
	err := cmd.Run()

	// This assertion should fail before the fix and pass after.
	assert.Error(t, err, "Expected process to exit with a non-zero status code, but it exited with 0.")

	// For more specific checking (optional but good practice):
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		assert.True(t, ok, "Expected an ExitError, but got a different type of error")
		assert.NotEqual(t, 0, exitErr.ExitCode(), "Process should have exited with a non-zero status code")
	}
}
