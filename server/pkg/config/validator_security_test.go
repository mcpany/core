// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestValidate_Security_VolumeMounts(t *testing.T) {
	// This test reproduces a security vulnerability where insecure volume mounts
	// (using ".." traversal) are allowed in the container environment configuration.
	// We first assert that it IS allowed (proving the issue), then we will fix it
	// and update the assertion.

	jsonConfig := `{
		"upstream_services": [
			{
				"name": "malicious-cmd-svc",
				"command_line_service": {
					"command": "echo hacked",
					"container_environment": {
						"image": "ubuntu",
						"volumes": {
							"../../../etc/passwd": "/target"
						}
					}
				}
			}
		]
	}`

	cfg := &configv1.McpAnyServerConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(jsonConfig), cfg))

	validationErrors := Validate(context.Background(), cfg, Server)

	// We expect validation errors now because the vulnerability is fixed
	require.NotEmpty(t, validationErrors, "Expected validation errors for insecure volume mount")
	assert.Contains(t, validationErrors[0].Error(), "is not a secure path")
	assert.Contains(t, validationErrors[0].Error(), "container environment volume host path")
}

func TestValidate_Security_ArbitraryExecutable_Blocked(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows for simplicity of permissions")
	}

	// Create an executable file in a temporary directory
	// Since we haven't configured IsAllowedPath to allow this temp dir, it should be blocked.
	tmpDir := t.TempDir()
	exePath := filepath.Join(tmpDir, "my_script.sh")
	err := os.WriteFile(exePath, []byte("#!/bin/sh\necho hello"), 0755)
	require.NoError(t, err)

	// Verify it is absolute
	require.True(t, filepath.IsAbs(exePath))

	// 1. New behavior: It should FAIL because the path is not in AllowedPaths or PATH
	err = validateCommandExists(exePath, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "is not allowed", "Should block execution of arbitrary file not in allowed paths")

	// 2. Verify non-existent file in blocked path also returns "not allowed" (or at least fails)
	// Actually, the check happens BEFORE existence check.
	// So probing /tmp/missing should ALSO fail with "not allowed".
	// This proves we don't leak existence!
	err = validateCommandExists(filepath.Join(tmpDir, "missing"), "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "is not allowed", "Should block probing of non-existent file in restricted path")

	// 3. Verify System Binary (in PATH) is allowed
	// Find 'ls'
	lsPath, err := exec.LookPath("ls")
	if err == nil && filepath.IsAbs(lsPath) {
		err = validateCommandExists(lsPath, "")
		require.NoError(t, err, "Should allow system binary in PATH")
	}
}
