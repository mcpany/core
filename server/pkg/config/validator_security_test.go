// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
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

func TestValidate_Security_MtlsPathTraversal(t *testing.T) {
	// Set allowed paths to a restricted directory
	allowedDir := t.TempDir()
	validation.SetAllowedPaths([]string{allowedDir})
	defer validation.SetAllowedPaths(nil)

	// Create a dummy file outside the allowed directory
	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "secret.pem")
	err := os.WriteFile(outsideFile, []byte("SECRET"), 0600)
	require.NoError(t, err)

	// Construct a config that uses the outside file in MTLS
	cfg := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("test-service"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: proto.String("http://localhost:8080"),
					},
				},
				UpstreamAuth: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_Mtls{
						Mtls: &configv1.MTLSAuth{
							ClientCertPath: proto.String(outsideFile), // Pointing to file OUTSIDE allowed paths
							ClientKeyPath:  proto.String(outsideFile),
						},
					},
				},
			},
		},
	}

	errs := Validate(context.Background(), cfg, Server)
	require.NotEmpty(t, errs, "Validation should fail for insecure path")
	assert.Contains(t, errs[0].Error(), "is not a secure path")

	// Test relative path traversal
	cfg.UpstreamServices[0].UpstreamAuth.GetMtls().ClientCertPath = proto.String("../secret.pem")
	errs = Validate(context.Background(), cfg, Server)
	require.NotEmpty(t, errs, "Validation should fail for relative path traversal")
	assert.Contains(t, errs[0].Error(), "is not a secure path")
}

func TestValidate_Security_GCPathTraversal(t *testing.T) {
	// Set allowed paths to a restricted directory
	allowedDir := t.TempDir()
	validation.SetAllowedPaths([]string{allowedDir})
	defer validation.SetAllowedPaths(nil)

	// Outside directory that should NOT be allowed
	outsideDir := t.TempDir()

	// Construct a config that uses the outside directory in GC
	cfg := &configv1.McpAnyServerConfig{
		GlobalSettings: &configv1.GlobalSettings{
			GcSettings: &configv1.GCSettings{
				Enabled:  proto.Bool(true),
				Interval: proto.String("1h"),
				Ttl:      proto.String("24h"),
				Paths:    []string{outsideDir},
			},
		},
	}

	// Validate should FAIL
	errs := Validate(context.Background(), cfg, Server)
	require.NotEmpty(t, errs, "GC Validation should fail for insecure path")
	assert.Contains(t, errs[0].Error(), "gc path")
	assert.Contains(t, errs[0].Error(), "is not secure")
}
