// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
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

func TestValidateSchema_Security_ExternalRefs(t *testing.T) {
	// Test LFI Protection
	t.Run("LFI_FileReference", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "secret.json")
		err := os.WriteFile(tmpFile, []byte(`{"type": "string"}`), 0644)
		require.NoError(t, err)

		schemaMap := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"foo": map[string]interface{}{
					"$ref": "file://" + tmpFile,
				},
			},
		}

		s, err := structpb.NewStruct(schemaMap)
		require.NoError(t, err)

		err = validateSchema(s)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "external schema references are disabled for security")
		assert.Contains(t, err.Error(), "file://")
	})

	// Test SSRF Protection
	t.Run("SSRF_HTTPReference", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"type": "string"}`))
		}))
		defer ts.Close()

		schemaMap := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"foo": map[string]interface{}{
					"$ref": ts.URL,
				},
			},
		}

		s, err := structpb.NewStruct(schemaMap)
		require.NoError(t, err)

		err = validateSchema(s)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "external schema references are disabled for security")
		assert.Contains(t, err.Error(), "http://")
	})
}
