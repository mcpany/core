// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"os"
	"path/filepath"
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

	cfg := configv1.McpAnyServerConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(jsonConfig), cfg))

	validationErrors := Validate(context.Background(), cfg, Server)

	// We expect validation errors now because the vulnerability is fixed
	require.NotEmpty(t, validationErrors, "Expected validation errors for insecure volume mount")
	assert.Contains(t, validationErrors[0].Error(), "is not a secure path")
	assert.Contains(t, validationErrors[0].Error(), "container environment volume host path")
}

func TestValidate_Security_StdioArgs_PathBypass(t *testing.T) {
	// This test ensures that script arguments for stdio connection must be within allowed paths.
	// Previously, validateStdioArgs only checked for existence, allowing execution of scripts outside CWD.

	// 1. Create a directory OUTSIDE the current working directory.
	tempDir, err := os.MkdirTemp("", "mcpany-repro-outside")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 2. Create a "malicious" script in that external directory.
	scriptPath := filepath.Join(tempDir, "evil_script.py")
	err = os.WriteFile(scriptPath, []byte("print('pwned')"), 0644)
	require.NoError(t, err)

	// 3. Configure a tool that uses python to execute this script.
	jsonConfig := `{
		"upstream_services": [
			{
				"name": "test-service-bypass",
				"mcp_service": {
					"stdio_connection": {
						"command": "python3",
						"args": ["` + scriptPath + `"]
					}
				}
			}
		]
	}`

	cfg := configv1.McpAnyServerConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(jsonConfig), cfg))

	// Mock execLookPath to ensure command validation passes (so we reach stdio arg validation)
	oldLookPath := execLookPath
	defer func() { execLookPath = oldLookPath }()
	execLookPath = func(file string) (string, error) {
		return "/usr/bin/python3", nil
	}

	// 4. Validate the config.
	// We expect validation errors because the script is outside CWD.
	errors := Validate(context.Background(), cfg, Server)

	assert.NotEmpty(t, errors, "Expected validation error for script outside CWD (vulnerability blocked)")
	if len(errors) > 0 {
		assert.Contains(t, errors[0].Error(), "file path")
		assert.Contains(t, errors[0].Error(), "is not allowed")
	}
}

func TestValidate_Security_StdioArgs_NoExtension_Bypass(t *testing.T) {
	// 1. Create a directory OUTSIDE the current working directory.
	tempDir, err := os.MkdirTemp("", "mcpany-repro-outside-noext")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 2. Create a "malicious" script in that external directory WITHOUT an extension.
	scriptPath := filepath.Join(tempDir, "evil_script")
	err = os.WriteFile(scriptPath, []byte("print('pwned')"), 0644)
	require.NoError(t, err)

	// 3. Configure a tool that uses python to execute this script.
	jsonConfig := `{
		"upstream_services": [
			{
				"name": "test-service-bypass-noext",
				"mcp_service": {
					"stdio_connection": {
						"command": "python3",
						"args": ["` + scriptPath + `"]
					}
				}
			}
		]
	}`

	cfg := configv1.McpAnyServerConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(jsonConfig), cfg))

	// Mock execLookPath to ensure command validation passes (so we reach stdio arg validation)
	oldLookPath := execLookPath
	defer func() { execLookPath = oldLookPath }()
	execLookPath = func(file string) (string, error) {
		return "/usr/bin/python3", nil
	}

	// 4. Validate the config.
	// We expect validation errors because the vulnerability is fixed.
	// The script is outside allowed paths, so it should be blocked even without an extension.
	errors := Validate(context.Background(), cfg, Server)

	require.NotEmpty(t, errors, "Expected validation errors for script outside CWD (vulnerability blocked)")
	if len(errors) > 0 {
		require.Contains(t, errors[0].Error(), "is not allowed")
	}
}
