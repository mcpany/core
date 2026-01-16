// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
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

func TestValidate_Security_CommandLineWorkingDirectory(t *testing.T) {
	// Test insecure path
	jsonConfig := `{
		"upstream_services": [
			{
				"name": "malicious-workdir-svc",
				"command_line_service": {
					"command": "ls",
					"working_directory": "../../../etc"
				}
			}
		]
	}`

	cfg := &configv1.McpAnyServerConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(jsonConfig), cfg))

	validationErrors := Validate(context.Background(), cfg, Server)

	require.NotEmpty(t, validationErrors, "Expected validation errors for insecure working directory")
	assert.Contains(t, validationErrors[0].Error(), "insecure working_directory")
}

func TestValidate_Security_FilesystemRootPaths(t *testing.T) {
	jsonConfig := `{
		"upstream_services": [
			{
				"name": "malicious-fs-svc",
				"filesystem_service": {
					"root_paths": {
                        "/virtual": "../../../etc"
                    }
				}
			}
		]
	}`

	cfg := &configv1.McpAnyServerConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(jsonConfig), cfg))

	validationErrors := Validate(context.Background(), cfg, Server)

	require.NotEmpty(t, validationErrors, "Expected validation errors for insecure root path")
	assert.Contains(t, validationErrors[0].Error(), "is not secure")
}

func TestValidate_VectorService(t *testing.T) {
	jsonConfig := `{
		"upstream_services": [
			{
				"name": "vector-svc",
				"vector_service": {
					"pinecone": {
                        "api_key": "",
                        "index_name": "idx"
                    }
				}
			}
		]
	}`

	cfg := &configv1.McpAnyServerConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(jsonConfig), cfg))

	validationErrors := Validate(context.Background(), cfg, Server)

	require.NotEmpty(t, validationErrors, "Expected validation errors for empty api key")
	assert.Contains(t, validationErrors[0].Error(), "pinecone api_key is empty")
}
