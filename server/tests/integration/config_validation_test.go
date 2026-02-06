// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConfigValidationFailures(t *testing.T) {
	root, err := GetProjectRoot()
	require.NoError(t, err)
	buildDir := GetBuildDir(t)

	mcpanyBinary := filepath.Join(buildDir, "bin/server")

	// Determine if root is server or repo to find tests
	var testsDir string
	if filepath.Base(root) == "server" {
		testsDir = filepath.Join(root, "tests", "integration")
	} else {
		testsDir = filepath.Join(root, "server", "tests", "integration")
	}

	tests := []struct {
		name          string
		configFile    string
		env           []string
		expectedError string
	}{
		{
			name:          "regex_fail_plaintext",
			configFile:    "testdata/test_regex_fail.yaml",
			expectedError: "secret value does not match validation regex",
		},
		{
			name:          "regex_fail_env_var",
			configFile:    "testdata/test_regex_env_fail.yaml",
			env:           []string{"TEST_ENV_VAL=invalid-key"},
			expectedError: "secret value does not match validation regex",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			absConfigFile := filepath.Join(testsDir, tc.configFile)

			// We use port 0 to avoid binding issues, but since validation fails early, ports might not even be bound.
			// NewManagedProcess requires arguments.
			// We use "run" command.
			args := []string{
				"run",
				"--config-path", absConfigFile,
				"--mcp-listen-address", "127.0.0.1:0", // Use port 0
				"--grpc-port", "127.0.0.1:0",
			}

			// Prepare Env
			env := []string{"MCPANY_LOG_LEVEL=debug", "MCPANY_ENABLE_FILE_CONFIG=true"}
			env = append(env, tc.env...)

			// Set label for logging
			mp := NewManagedProcess(t, "ConfigValidation-"+tc.name, mcpanyBinary, args, env)
			mp.IgnoreExitStatusOne = true // We expect exit status 1

			err := mp.Start()
			require.NoError(t, err, "Process should start")

			// Wait for exit
			select {
			case <-mp.waitDone:
				// Process exited
				if mp.cmd.ProcessState != nil && mp.cmd.ProcessState.Success() {
					t.Fatalf("Process exited successfully but should have failed with validation error. Output:\n%s", mp.StdoutString())
				}
			case <-time.After(10 * time.Second):
				mp.Stop()
				t.Fatalf("Process did not exit in time. It should have failed validation.")
			}

			// Check output for error message
			// Validation errors are usually logged to Stderr or Stdout depending on logger config.
			// Using CombinedOutput logic via mp.stdout/stderr
			output := mp.StdoutString() + mp.StderrString()
			require.Contains(t, output, tc.expectedError, "Output should contain validation error")
		})
	}
}
